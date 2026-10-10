package factory

import (
	"github.com/damolaoladipo/pricebase/internal/mappers"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

var UserService *services.UserService
var AuthService *services.AuthService
var SystemService *services.SystemService

func InitServices(db *pgxpool.Pool) {

	userRepository := repository.NewUserRepository(db)
	roleRepository := repository.NewRoleRepository(db)
	permissionRepository := repository.NewPermissionRepository(db)
	userMapper := mappers.NewUserMapper()

	SystemService = services.NewSystemService()
	UserService = services.NewUserService(userRepository, SystemService, roleRepository)
	AuthService = services.NewAuthService(roleRepository, permissionRepository, userRepository, userMapper)
}
