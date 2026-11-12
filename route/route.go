package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gofiber/swagger"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

// Dependencies mengumpulkan semua yang dibutuhkan Register.
// Ditulis sebagai struct agar penambahan service berikutnya
// tidak mengubah tanda tangan function.
type Dependencies struct {
	Pool         *pgxpool.Pool
	JWT          *helper.JWTManager
	Permissions *helper.PermissionSet
	StudentSvc   *service.StudentService
	AuthSvc      *service.AuthService
	SupervisorService *service.SupervisorService
	PortfolioService *service.PortfolioService
	Mongo             *mongo.Client
	CourseService       *service.CourseService
	ReviewService       *service.ReviewService
	NotificationService *service.NotificationService
	EnableSwagger bool
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	if deps.EnableSwagger {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}
	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))
	api.Get("/portfolios/stats/technologies", 
		middleware.RequireAuth(deps.JWT), 
		deps.PortfolioService.StatsTechnologies,
	)

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthSvc.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthSvc.Login)
	auth.Post("/refresh", deps.AuthSvc.Refresh)
	auth.Post("/logout", deps.AuthSvc.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthSvc.Me)

	// --- supervisor ---
	api.Post("/supervisions/switch", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.SupervisorService.Switch)

	// --- wajib membawa access token ---
	students := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	students.Get("/",
		middleware.RequirePermission(deps.Permissions, "student:list"),
		deps.StudentSvc.List)
	students.Get("/:id", deps.StudentSvc.Get)
	students.Post("/",
		middleware.RequirePermission(deps.Permissions, "student:create"),
		deps.StudentSvc.Create)
	students.Put("/:id", deps.StudentSvc.Replace)
	students.Patch("/:id", deps.StudentSvc.Patch)
	students.Delete("/:id",
		middleware.RequirePermission(deps.Permissions, "student:delete"),
		deps.StudentSvc.Delete)
	sup := api.Group("/supervisors", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	sup.Post("/:id/assign", middleware.RequirePermission(deps.Permissions, "student:create"), deps.SupervisorService.Assign)
	sup.Patch("/:id", middleware.RequirePermission(deps.Permissions, "student:update:any"), deps.SupervisorService.UpdateQuota)
	portfolios := api.Group("/students/:id/portfolio", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	portfolios.Get("", deps.PortfolioService.Get)
	portfolios.Put("", deps.PortfolioService.Upsert)
	portfolios.Post("/projects", deps.PortfolioService.AddProject)
	portfolios.Delete("/projects/:projectId", deps.PortfolioService.DeleteProject)

	// --- kelas dan KRS ---
	courses := api.Group("/courses", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(deps.Permissions, "course:manage"), deps.CourseService.UpdateQuota)
	courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
	courses.Post("/:id/reviews", deps.ReviewService.Create)

	api.Post("/enrollments/swap", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), deps.CourseService.Swap)

	// --- notifikasi (Modul 11) ---
	notifications := api.Group("/notifications", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	notifications.Get("/mine", deps.NotificationService.Mine)

	admin := api.Group("/admin", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	admin.Get("/queue", middleware.RequirePermission(deps.Permissions, "queue:monitor"), deps.NotificationService.Monitor)

}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}