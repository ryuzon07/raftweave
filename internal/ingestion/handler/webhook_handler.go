package handler

import (
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
	"github.com/raftweave/raftweave/internal/ingestion/usecase"
)

type WebhookHandler struct {
	logger         *zap.Logger
	processWebhook *usecase.ProcessWebhookUseCase
}

func NewWebhookHandler(logger *zap.Logger, processWebhook *usecase.ProcessWebhookUseCase) *WebhookHandler {
	return &WebhookHandler{
		logger:         logger,
		processWebhook: processWebhook,
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read webhook body", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Extract standard webhook headers to map to domain model
	// This supports multiple providers generically
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	// Detect provider (e.g., GitHub, GitLab)
	provider := domain.WebhookProvider("unknown")
	var signature string

	if sig := r.Header.Get("X-Hub-Signature-256"); sig != "" {
		provider = "github"
		signature = sig
	} else if sig := r.Header.Get("X-Gitlab-Token"); sig != "" {
		provider = "gitlab"
		signature = sig
	} else {
		h.logger.Warn("missing webhook signature header")
		http.Error(w, "Missing Signature", http.StatusUnauthorized)
		return
	}

	in := usecase.ProcessWebhookInput{
		Provider:   provider,
		RawPayload: payload,
		Signature:  signature,
		Headers:    headers,
	}

	out, err := h.processWebhook.Execute(r.Context(), in)
	if err != nil {
		h.logger.Error("webhook processing failed", zap.Error(err))

		if err == domain.ErrInvalidSignature {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
		if err == domain.ErrWorkloadNotFound {
			http.Error(w, "Workload not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("successfully accepted webhook",
		zap.String("job_id", out.JobID),
		zap.String("workload", out.WorkloadName),
		zap.String("commit", out.CommitSHA),
	)

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("Webhook accepted"))
}