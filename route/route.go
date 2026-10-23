package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

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
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

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
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}