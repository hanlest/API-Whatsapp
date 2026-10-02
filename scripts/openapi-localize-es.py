#!/usr/bin/env python3
"""Localiza docs/openapi.yaml al español (summaries, tags, respuestas HTTP comunes)."""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OPENAPI = ROOT / "docs" / "openapi.yaml"
EMBED = ROOT / "src" / "ui" / "rest" / "swagger" / "openapi.yaml"

INFO = '''openapi: "3.0.0"
info:
  title: WhatsApp API MultiDevice
  version: 9.0.0
  description: |
    API HTTP para operar WhatsApp Web (multidispositivo): envío de mensajes, chats, grupos,
    dispositivos, webhooks y opcionalmente Chatwoot.

    **Panel web:** el dashboard [gowa-ui](https://github.com/aldinokemal/gowa-ui) se descarga
    al arrancar y se sirve en `/` (o un aviso JSON si `APP_UI_ENABLED=false`).

    **Documentación interactiva:** Swagger UI en `/docs/` (misma autenticación Basic Auth que la API).

    **Alcance por dispositivo:**
    - Enviá el header `X-Device-Id` en todas las rutas que operan sobre una cuenta concreta.
    - También podés usar el query `device_id` donde la operación lo admita.
    - WebSocket: `/ws?device_id=<id>`.
    - Con Basic Auth activo, los navegadores no pueden enviar `Authorization` en WebSocket;
      podés usar `&authorization=<base64(usuario:contraseña)>` en la URL (usá TLS en producción).
'''

TAGS = '''tags:
  - name: app
    description: Conexión inicial, sesión y metadatos del servidor
  - name: device
    description: Alta, login y webhooks por cuenta (multidispositivo)
  - name: user
    description: Perfil, contactos, privacidad y listas de la cuenta
  - name: send
    description: Envío de mensajes y presencia (texto, medios, ubicación, encuestas, etc.)
  - name: message
    description: Acciones sobre mensajes (reaccionar, editar, borrar, reenviar, descargar medios)
  - name: call
    description: Llamadas entrantes (rechazo automático o manual)
  - name: chat
    description: Listado de chats, historial y ajustes de conversación
  - name: group
    description: Creación, miembros, permisos e invitaciones de grupos
  - name: newsletter
    description: Canales / newsletters (seguir, mensajes y medios)
  - name: chatwoot
    description: Integración con Chatwoot (sync, webhooks y config por dispositivo)
  - name: media-storage
    description: Almacenamiento de medios (local o MiniCloud) y proxy de descarga
  - name: docs
    description: Especificación OpenAPI y Swagger UI
'''

SUMMARY_MAP = {
    "Health check": "Comprobar estado del servicio",
    "Login to whatsapp server": "Iniciar sesión (código QR)",
    "Login with pairing code": "Iniciar sesión con código de emparejamiento",
    "Get pending passkey pairing state": "Estado del emparejamiento passkey pendiente",
    "Submit the WebAuthn assertion for passkey pairing": "Enviar aserción WebAuthn (passkey)",
    "Confirm the passkey pairing code": "Confirmar código de emparejamiento passkey",
    "Logout the active device (keeps the device slot)": "Cerrar sesión (conserva el slot del dispositivo)",
    "Reconnecting to whatsapp server": "Reconectar con los servidores de WhatsApp",
    "Get list connected devices": "Listar dispositivos conectados (legacy /app)",
    "Get connection status": "Estado de conexión de la sesión activa",
    "Get server metadata": "Metadatos del servidor (versión, límites de medios)",
    "List all devices": "Listar dispositivos registrados",
    "Add a new device": "Registrar un nuevo dispositivo / cuenta",
    "Get device info": "Información de un dispositivo",
    "Remove a device": "Eliminar un dispositivo y sus datos asociados",
    "Login a device with QR code": "Login del dispositivo con código QR",
    "Login a device with pairing code": "Login del dispositivo con código numérico",
    "Logout device": "Cerrar sesión del dispositivo",
    "Reconnect device": "Forzar reconexión del dispositivo",
    "Get device connection status": "Estado de conexión del dispositivo",
    "Get device webhook configuration": "Obtener webhook configurado del dispositivo",
    "Set device webhook configuration": "Configurar webhook del dispositivo",
    "User Info": "Información del usuario de WhatsApp",
    "User Avatar": "Obtener avatar del usuario",
    "User Change Avatar": "Cambiar avatar de la cuenta",
    "User Change Push Name": "Cambiar nombre visible (push name)",
    "User My Privacy Setting": "Consultar ajustes de privacidad",
    "User My List Groups": "Listar mis grupos",
    "Get list of user contacts": "Listar contactos de la agenda",
    "Check if user is on WhatsApp": "Comprobar si un número tiene WhatsApp",
    "Get Business Profile Information": "Perfil comercial (Business)",
    "Send Message": "Enviar mensaje de texto",
    "Send Image": "Enviar imagen",
    "Send Audio": "Enviar audio o nota de voz",
    "Send File": "Enviar documento / archivo",
    "Send Sticker": "Enviar sticker",
    "Send Video": "Enviar video",
    "Send Contact": "Enviar contacto (vCard)",
    "Send Link": "Enviar enlace con vista previa",
    "Send Location": "Enviar ubicación",
    "Send Poll / Vote": "Enviar encuesta",
    "Send presence status": "Publicar estado de presencia global",
    "Send chat presence (typing indicator)": "Indicador de escritura en un chat",
    "Revoke Message": "Revocar mensaje para todos",
    "Delete Message": "Eliminar mensaje (solo en este dispositivo)",
    "Send reaction to message": "Reaccionar a un mensaje",
    "Edit message by message ID before 15 minutes": "Editar mensaje (ventana ~15 min)",
    "Mark as read message": "Marcar mensaje como leído",
    "Mark an audio message as played": "Marcar audio como reproducido",
    "Star message": "Destacar mensaje",
    "Unstar message": "Quitar destacado",
    "Forward Message": "Reenviar mensaje a otro chat",
    "Download media from message": "Descargar medios de un mensaje almacenado",
    "Reject an incoming call": "Rechazar llamada entrante",
    "Get list of chats": "Listar conversaciones",
    "Get messages from a specific chat": "Historial de mensajes de un chat",
    "Pin or unpin a chat": "Fijar o desfijar chat",
    "Set disappearing messages timer": "Configurar mensajes temporales",
    "Archive or unarchive a chat": "Archivar o desarchivar chat",
    "Group Info": "Información del grupo",
    "Create group and add participant": "Crear grupo e invitar participantes",
    "Get list of participants in a group": "Listar participantes del grupo",
    "Adding more participants to group": "Añadir participantes al grupo",
    "Export group participants as CSV": "Exportar participantes (CSV)",
    "Remove participants from group": "Expulsar participantes",
    "Promote participants to admin": "Promover a administradores",
    "Demote participants to member": "Degradar administradores a miembros",
    "Join group with link": "Unirse al grupo con enlace",
    "Get group information from invitation link": "Vista previa del grupo desde enlace",
    "Get list of participant requests to join group": "Solicitudes pendientes de ingreso",
    "Approve participant request to join group": "Aprobar solicitud de ingreso",
    "Reject participant request to join group": "Rechazar solicitud de ingreso",
    "Leave group": "Abandonar grupo",
    "Set group photo": "Cambiar foto del grupo",
    "Set group name": "Cambiar nombre del grupo",
    "Set group locked status": "Bloquear edición del grupo (solo admins)",
    "Set group announce mode": "Modo anuncios (solo admins escriben)",
    "Set group topic": "Cambiar descripción / asunto del grupo",
    "Group Invite Link": "Obtener o regenerar enlace de invitación",
    "Unfollow newsletter": "Dejar de seguir un canal / newsletter",
    "Get latest newsletter (channel) messages": "Últimos mensajes de un canal",
    "Download media from a newsletter message": "Descargar medios de un mensaje de canal",
    "Sync message history to Chatwoot": "Sincronizar historial hacia Chatwoot",
    "Get Chatwoot sync progress": "Progreso de la sincronización Chatwoot",
    "Chatwoot webhook endpoint": "Webhook global de Chatwoot (entrante)",
    "Per-device Chatwoot webhook endpoint": "Webhook Chatwoot por dispositivo",
    "List per-device Chatwoot configs": "Listar configs Chatwoot por dispositivo",
    "Get a device's Chatwoot config": "Obtener config Chatwoot del dispositivo",
    "Create or update a device's Chatwoot config": "Crear o actualizar config Chatwoot",
    "Delete a device's Chatwoot config": "Eliminar config Chatwoot del dispositivo",
}

HTTP_DESC = {
    "OK": "Operación exitosa",
    "Bad Request": "Solicitud inválida",
    "Internal Server Error": "Error interno del servidor",
    "Service healthy": "Servicio disponible",
    "Service unavailable": "Servicio no disponible",
    "Media downloaded successfully": "Medios descargados correctamente",
    "Message lookup or media download failed": "No se pudo obtener el mensaje o descargar el medio",
}

EXTRA_PATHS = '''
  /docs/:
    get:
      operationId: swaggerUI
      tags:
        - docs
      summary: Documentación interactiva (Swagger UI)
      description: |
        Interfaz Swagger para explorar y probar la API. Requiere la misma autenticación
        Basic Auth que el resto de endpoints (excepto rutas públicas como `/health`).
      responses:
        '200':
          description: Página HTML de Swagger UI
          content:
            text/html:
              schema:
                type: string
  /docs/openapi.yaml:
    get:
      operationId: openAPISpec
      tags:
        - docs
      summary: Especificación OpenAPI (YAML)
      description: Contrato OpenAPI 3.0 embebido/servido por el servidor; la URL base del servidor se ajusta al host de la petición.
      responses:
        '200':
          description: Documento OpenAPI en YAML
          content:
            application/yaml:
              schema:
                type: string
        '404':
          description: Especificación no disponible
  /media-storage/configs:
    get:
      operationId: listMediaStorageConfigs
      tags:
        - media-storage
      summary: Listar configuración de almacenamiento de medios
      description: Devuelve bucket, carpeta y driver efectivo (local o MiniCloud) por dispositivo.
      security:
        - basicAuth: []
      responses:
        '200':
          description: Operación exitosa
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GenericResponse'
  /devices/{device_id}/media-storage:
    get:
      operationId: getDeviceMediaStorage
      tags:
        - media-storage
      summary: Obtener almacenamiento de medios del dispositivo
      parameters:
        - name: device_id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Operación exitosa
        '404':
          description: Dispositivo o configuración no encontrados
    put:
      operationId: upsertDeviceMediaStorage
      tags:
        - media-storage
      summary: Configurar bucket/carpeta MiniCloud por dispositivo
      description: |
        Permite bucket dedicado o compartido y nombre de carpeta dentro del bucket.
        Requiere `MEDIA_STORAGE_DRIVER=minicloud` y `MINICLOUD_BASE_URL` configurados.
      parameters:
        - name: device_id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                bucket:
                  type: string
                folder_name:
                  type: string
                enabled:
                  type: boolean
      responses:
        '200':
          description: Configuración guardada
    delete:
      operationId: deleteDeviceMediaStorage
      tags:
        - media-storage
      summary: Quitar configuración de medios del dispositivo
      parameters:
        - name: device_id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Configuración eliminada
  /devices/{device_id}/media/{object_key}:
    get:
      operationId: proxyDeviceMedia
      tags:
        - media-storage
      summary: Descargar medio almacenado (proxy)
      description: |
        Sirve archivos referenciados como `mc:bucket/clave` o rutas locales bajo `statics/media`,
        según `MEDIA_PROXY_ENABLED` y el driver configurado. Requiere Basic Auth.
      parameters:
        - name: device_id
          in: path
          required: true
          schema:
            type: string
        - name: object_key
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Contenido binario del medio
        '404':
          description: Medio no encontrado
'''


def main() -> None:
    text = OPENAPI.read_text(encoding="utf-8")

    # info + tags block (until security:)
    text = re.sub(
        r"openapi:.*?^security:",
        INFO + TAGS + "security:",
        text,
        count=1,
        flags=re.MULTILINE | re.DOTALL,
    )

    for en, es in SUMMARY_MAP.items():
        text = text.replace(f"summary: {en}", f"summary: {es}")

    for en, es in HTTP_DESC.items():
        text = text.replace(f"description: {en}", f"description: {es}")

    # Parámetros globales
    text = text.replace(
        """      description: |
        Device identifier for multi-device support. Required when multiple devices are registered.
        If only one device is registered, it will be used as the default.
        Can also be provided as `device_id` query parameter.""",
        """      description: |
        Identificador del dispositivo/cuenta. Obligatorio si hay varios dispositivos registrados;
        con un solo dispositivo se usa por defecto. También podés pasar `device_id` como query.""",
    )
    text = text.replace(
        "description: Device ID (or WhatsApp JID) whose Chatwoot config is addressed",
        "description: ID del dispositivo (o JID de WhatsApp) cuya config Chatwoot se consulta o modifica",
    )

    # Ejemplos de message en schemas (alineados con controladores en español)
    message_examples = {
        "Success get chat list": "Lista de chats obtenida correctamente",
        "Success get chat messages": "Mensajes del chat obtenidos correctamente",
        "Success get list groups": "Lista de grupos obtenida correctamente",
        "Success get group info from link": "Información del grupo obtenida desde el enlace",
        "Success get group info": "Información del grupo obtenida correctamente",
        "Success get group invite link": "Enlace de invitación obtenido correctamente",
        "Chat pinned successfully": "Chat fijado correctamente",
        "Disappearing messages set to 86400 seconds": "Mensajes temporales configurados a 86400 segundos",
        "Chat archived successfully": "Chat archivado correctamente",
        "History sync initiated in background": "Sincronización de historial iniciada en segundo plano",
        "Fetch app info success": "Metadatos del servidor obtenidos correctamente",
        "Login success": "Inicio de sesión iniciado correctamente",
    }
    for en, es in message_examples.items():
        text = text.replace(f"example: {en}", f"example: {es}")

    if "/docs/:" not in text:
        text = text.replace("\ncomponents:", EXTRA_PATHS + "\ncomponents:", 1)

    OPENAPI.write_text(text, encoding="utf-8")
    EMBED.parent.mkdir(parents=True, exist_ok=True)
    EMBED.write_text(text, encoding="utf-8")
    print(f"Updated {OPENAPI} and {EMBED}")


if __name__ == "__main__":
    main()
