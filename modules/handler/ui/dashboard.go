package ui

import (
	"context"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/service"
	"coreum_processor/modules/user"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func PageDashboard(ctx context.Context, userService *user.Service, processing *service.ProcessingService) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		userStore, err := internal.GetUserStore(request.Context())
		if err != nil {
			PageLanding(writer, request, params)
		} else if user.IsSysAdmin(userStore.Access) {
			PageRequestsAdmin(ctx, userService, processing)(writer, request, params)
		} else if user.IsOnboarding(userStore.Access) {
			PageWizardMerchant(ctx, userService)(writer, request, params)
		} else if user.IsOnboarded(userStore.Access) {
			PageMerchantTransaction(ctx, processing)(writer, request, params)
		}
		return
	}
}
