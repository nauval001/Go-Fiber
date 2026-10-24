package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
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
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

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