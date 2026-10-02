package rest

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	domainGroup "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/group"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"go.mau.fi/whatsmeow"
)

type Group struct {
	Service domainGroup.IGroupUsecase
}

func InitRestGroup(app fiber.Router, service domainGroup.IGroupUsecase) Group {
	rest := Group{Service: service}
	app.Post("/group", rest.CreateGroup)
	app.Post("/group/join-with-link", rest.JoinGroupWithLink)
	app.Get("/group/info-from-link", rest.GetGroupInfoFromLink)
	app.Get("/group/info", rest.GroupInfo)
	app.Post("/group/leave", rest.LeaveGroup)
	app.Get("/group/participants", rest.ListParticipants)
	app.Get("/group/participants/export", rest.ExportParticipants)
	app.Post("/group/participants", rest.AddParticipants)
	app.Post("/group/participants/remove", rest.DeleteParticipants)
	app.Post("/group/participants/promote", rest.PromoteParticipants)
	app.Post("/group/participants/demote", rest.DemoteParticipants)
	app.Get("/group/participant-requests", rest.ListParticipantRequests)
	app.Post("/group/participant-requests/approve", rest.ApproveParticipantRequests)
	app.Post("/group/participant-requests/reject", rest.RejectParticipantRequests)
	app.Post("/group/photo", rest.SetGroupPhoto)
	app.Post("/group/name", rest.SetGroupName)
	app.Post("/group/locked", rest.SetGroupLocked)
	app.Post("/group/announce", rest.SetGroupAnnounce)
	app.Post("/group/topic", rest.SetGroupTopic)
	app.Get("/group/invite-link", rest.GetGroupInviteLink)
	return rest
}

func (controller *Group) JoinGroupWithLink(c fiber.Ctx) error {
	var request domainGroup.JoinGroupWithLinkRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	response, err := controller.Service.JoinGroupWithLink(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Te uniste al grupo correctamente",
		Results: map[string]string{
			"group_id": response,
		},
	})
}

func (controller *Group) GetGroupInfoFromLink(c fiber.Ctx) error {
	var request domainGroup.GetGroupInfoFromLinkRequest
	err := c.Bind().Query(&request)
	utils.PanicIfNeeded(err)

	response, err := controller.Service.GetGroupInfoFromLink(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Información del grupo obtenida desde el enlace",
		Results: response,
	})
}

func (controller *Group) LeaveGroup(c fiber.Ctx) error {
	var request domainGroup.LeaveGroupRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	err = controller.Service.LeaveGroup(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Abandonaste el grupo correctamente",
	})
}

func (controller *Group) CreateGroup(c fiber.Ctx) error {
	var request domainGroup.CreateGroupRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	groupID, err := controller.Service.CreateGroup(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Grupo creado correctamente (id %s)", groupID),
		Results: map[string]string{
			"group_id": groupID,
		},
	})
}

func (controller *Group) ListParticipants(c fiber.Ctx) error {
	var request domainGroup.GetGroupParticipantsRequest
	err := c.Bind().Query(&request)
	utils.PanicIfNeeded(err)

	if request.GroupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  400,
			Code:    "INVALID_GROUP_ID",
			Message: "El ID del grupo no puede estar vacío",
		})
	}

	utils.SanitizePhone(&request.GroupID)

	result, err := controller.Service.GetGroupParticipants(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Participantes del grupo obtenidos correctamente",
		Results: result,
	})
}

func (controller *Group) ExportParticipants(c fiber.Ctx) error {
	var request domainGroup.GetGroupParticipantsRequest
	err := c.Bind().Query(&request)
	utils.PanicIfNeeded(err)

	if request.GroupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  400,
			Code:    "INVALID_GROUP_ID",
			Message: "El ID del grupo no puede estar vacío",
		})
	}

	utils.SanitizePhone(&request.GroupID)

	result, err := controller.Service.GetGroupParticipants(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	utils.PanicIfNeeded(writer.Write([]string{"participant_jid", "phone_number", "lid", "display_name", "role"}))

	for _, participant := range result.Participants {
		role := "member"
		if participant.IsSuperAdmin {
			role = "super_admin"
		} else if participant.IsAdmin {
			role = "admin"
		}

		record := []string{
			participant.JID,
			participant.PhoneNumber,
			participant.LID,
			participant.DisplayName,
			role,
		}

		utils.PanicIfNeeded(writer.Write(record))
	}

	writer.Flush()
	utils.PanicIfNeeded(writer.Error())

	fileName := fmt.Sprintf("group-%s-participants.csv", strings.ReplaceAll(result.GroupID, "@", "_"))

	c.Type("text/csv; charset=utf-8")
	c.Attachment(fileName)

	return c.Send(buffer.Bytes())
}

func (controller *Group) AddParticipants(c fiber.Ctx) error {
	return controller.manageParticipants(c, whatsmeow.ParticipantChangeAdd, "Participantes añadidos correctamente")
}

func (controller *Group) DeleteParticipants(c fiber.Ctx) error {
	return controller.manageParticipants(c, whatsmeow.ParticipantChangeRemove, "Participantes eliminados correctamente")
}

func (controller *Group) PromoteParticipants(c fiber.Ctx) error {
	return controller.manageParticipants(c, whatsmeow.ParticipantChangePromote, "Participantes promovidos a administradores")
}

func (controller *Group) DemoteParticipants(c fiber.Ctx) error {
	return controller.manageParticipants(c, whatsmeow.ParticipantChangeDemote, "Administradores degradados a miembros")
}

func (controller *Group) ListParticipantRequests(c fiber.Ctx) error {
	var request domainGroup.GetGroupRequestParticipantsRequest
	err := c.Bind().Query(&request)
	utils.PanicIfNeeded(err)

	if request.GroupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  400,
			Code:    "INVALID_GROUP_ID",
			Message: "El ID del grupo no puede estar vacío",
		})
	}

	utils.SanitizePhone(&request.GroupID)

	result, err := controller.Service.GetGroupRequestParticipants(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Solicitudes de ingreso listadas correctamente",
		Results: result,
	})
}

func (controller *Group) ApproveParticipantRequests(c fiber.Ctx) error {
	return controller.handleRequestedParticipants(c, whatsmeow.ParticipantChangeApprove, "Solicitudes de ingreso aprobadas")
}

func (controller *Group) RejectParticipantRequests(c fiber.Ctx) error {
	return controller.handleRequestedParticipants(c, whatsmeow.ParticipantChangeReject, "Solicitudes de ingreso rechazadas")
}

// Generalized participant management handler
func (controller *Group) manageParticipants(c fiber.Ctx, action whatsmeow.ParticipantChange, successMsg string) error {
	var request domainGroup.ParticipantRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)
	utils.SanitizePhone(&request.GroupID)
	request.Action = action
	result, err := controller.Service.ManageParticipant(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: successMsg,
		Results: result,
	})
}

// Generalized requested participants handler
func (controller *Group) handleRequestedParticipants(c fiber.Ctx, action whatsmeow.ParticipantRequestChange, successMsg string) error {
	var request domainGroup.GroupRequestParticipantsRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)
	utils.SanitizePhone(&request.GroupID)
	request.Action = action
	result, err := controller.Service.ManageGroupRequestParticipants(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: successMsg,
		Results: result,
	})
}

func (controller *Group) SetGroupPhoto(c fiber.Ctx) error {
	var request domainGroup.SetGroupPhotoRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	file, err := c.FormFile("photo")
	if err == nil {
		logrus.Printf("INFO: Received group photo - Filename: %s, Size: %d bytes, ContentType: %s",
			file.Filename, file.Size, file.Header.Get("Content-Type"))

		// Basic validation only - processing will be done in usecase
		if err := utils.ValidateGroupPhotoFormat(file); err != nil {
			logrus.Printf("ERROR: Group photo validation failed - %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
				Status:  400,
				Code:    "INVALID_IMAGE_FORMAT",
				Message: fmt.Sprintf("Validación de imagen fallida: %v", err),
			})
		}

		request.Photo = file
	} else {
		logrus.Printf("DEBUG: No photo file provided - Error: %v", err)
	}

	pictureID, err := controller.Service.SetGroupPhoto(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	if err != nil {
		logrus.Printf("ERROR: WhatsApp service failed to set group photo - %v", err)
	}
	utils.PanicIfNeeded(err)

	message := "Foto del grupo actualizada correctamente"
	if request.Photo == nil {
		message = "Foto del grupo eliminada correctamente"
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: message,
		Results: domainGroup.SetGroupPhotoResponse{
			PictureID: pictureID,
			Message:   message,
		},
	})
}

func (controller *Group) SetGroupName(c fiber.Ctx) error {
	var request domainGroup.SetGroupNameRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	err = controller.Service.SetGroupName(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Nombre del grupo actualizado a '%s'", request.Name),
	})
}

func (controller *Group) SetGroupLocked(c fiber.Ctx) error {
	var request domainGroup.SetGroupLockedRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	err = controller.Service.SetGroupLocked(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	message := "Grupo desbloqueado (miembros pueden editar info)"
	if request.Locked {
		message = "Grupo bloqueado (solo administradores editan info)"
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: message,
	})
}

func (controller *Group) SetGroupAnnounce(c fiber.Ctx) error {
	var request domainGroup.SetGroupAnnounceRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	err = controller.Service.SetGroupAnnounce(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	message := "Modo anuncios desactivado"
	if request.Announce {
		message = "Modo anuncios activado (solo admins envían mensajes)"
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: message,
	})
}

func (controller *Group) SetGroupTopic(c fiber.Ctx) error {
	var request domainGroup.SetGroupTopicRequest
	err := c.Bind().Body(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	err = controller.Service.SetGroupTopic(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	message := "Descripción del grupo actualizada"
	if request.Topic == "" {
		message = "Descripción del grupo eliminada"
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: message,
	})
}

// GroupInfo handles the /group/info endpoint to fetch group information
func (controller *Group) GroupInfo(c fiber.Ctx) error {
	var request domainGroup.GroupInfoRequest
	err := c.Bind().Query(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	response, err := controller.Service.GroupInfo(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Información del grupo obtenida correctamente",
		Results: response.Data,
	})
}

func (controller *Group) GetGroupInviteLink(c fiber.Ctx) error {
	var request domainGroup.GetGroupInviteLinkRequest
	err := c.Bind().Query(&request)
	utils.PanicIfNeeded(err)

	utils.SanitizePhone(&request.GroupID)

	response, err := controller.Service.GetGroupInviteLink(whatsapp.ContextWithDevice(c.Context(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Enlace de invitación obtenido correctamente",
		Results: response,
	})
}
