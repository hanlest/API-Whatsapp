package error

import "net/http"

type WaSessionNotReadyError string

func (e WaSessionNotReadyError) Error() string {
	return string(e)
}

func (e WaSessionNotReadyError) ErrCode() string {
	return "WA_SESSION_NOT_READY"
}

func (e WaSessionNotReadyError) StatusCode() int {
	return http.StatusServiceUnavailable
}

const ErrWaSessionNotReady = WaSessionNotReadyError("WhatsApp session is not ready")
