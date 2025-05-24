package user

import (
	"github.com/tapas/go-user-service/internal/user/handler"
	"github.com/tapas/go-user-service/internal/user/repository"
	"github.com/tapas/go-user-service/internal/user/service"
	"gorm.io/gorm"
)

// ProvideUserRepository provides a UserRepository.
func ProvideUserRepository(db *gorm.DB) repository.UserRepository {
	return repository.NewUserRepository(db)
}

// ProvideUserService provides a UserService.
func ProvideUserService(userRepo repository.UserRepository) service.UserService {
	return service.NewUserService(userRepo)
}

// ProvideUserHandler provides a UserHandler.
func ProvideUserHandler(userService service.UserService) handler.UserHandler {
	return handler.NewUserHandler(userService)
}
