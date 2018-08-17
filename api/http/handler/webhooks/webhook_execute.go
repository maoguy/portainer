package webhooks

import (
	"context"
	"net/http"

	dockertypes "github.com/docker/docker/api/types"
	portainer "github.com/portainer/portainer"
	docker "github.com/portainer/portainer/docker"
	httperror "github.com/portainer/portainer/http/error"
	"github.com/portainer/portainer/http/request"
	"github.com/portainer/portainer/http/response"
)

// Acts on a passed in token UUID to restart the docker service
func (handler *Handler) webhookExecute(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	webhookToken, err := request.RetrieveRouteVariableValue(r, "token")

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Invalid service id parameter", err}
	}

	webhook, err := handler.WebhookService.WebhookByToken(webhookToken)
	if err != nil && err != portainer.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve webhook from the database", err}
	}

	serviceID := webhook.ServiceID
	endpointID := webhook.EndpointID

	endpoint, err := handler.EndpointService.Endpoint(portainer.EndpointID(endpointID))
	if err == portainer.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an endpoint with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an endpoint with the specified identifier inside the database", err}
	}

	dockerFactory := docker.NewClientFactory(handler.SignatureService)
	dockerClient, err := dockerFactory.CreateClient(endpoint)

	service, _, err := dockerClient.ServiceInspectWithRaw(context.Background(), serviceID, dockertypes.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error looking up service", err}
	}

	resp, err := dockerClient.ServiceUpdate(context.Background(), serviceID, service.Version, service.Spec, dockertypes.ServiceUpdateOptions{QueryRegistry: true})

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error updating service", err}
	}
	if resp.Warnings != nil {
		//Log warnings
	}
	return response.Empty(w)

}