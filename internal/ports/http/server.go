package http

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"service/internal/entities"
	"service/pkg/dto"
)

const (
	userIDURLParam = "user_id"
)

type Server struct {
	router *chi.Mux
	svc    BalanceService
	port   int
	log    *zap.SugaredLogger
}

func NewServer(log *zap.SugaredLogger, svc BalanceService, port int) (*Server, error) {
	// validation errors

	router := chi.NewRouter()

	server := &Server{
		router: router,
		svc:    svc,
		port:   port,
		log:    log,
	}

	basePath := "/api/v1"

	router.Route(basePath, func(r chi.Router) {
		r.Get(fmt.Sprintf("/{%s}", userIDURLParam), server.GetUserBalance)
		r.Post(fmt.Sprintf("/{%s}", userIDURLParam), server.CreditBalance)
	})

	return server, nil
}

func (s *Server) Run() {
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router)
}

func (s *Server) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := chi.URLParam(r, userIDURLParam)
	if userID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty balance id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	balance, err := s.svc.GetUserBalance(ctx, userID)
	if errors.Is(err, entities.ErrNotFound) {
		s.log.Error(err)
		s.writeError(w, err, http.StatusNotFound)
		return
	}
	if err != nil {
		s.log.Error(err)
		s.writeError(w, entities.ErrInternal, http.StatusInternalServerError)
		return
	}

	response := &dto.Balance{
		UserID:   balance.UserID(),
		Currency: int(balance.Currency()),
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) CreditBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := chi.URLParam(r, userIDURLParam)
	if userID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty balance id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	request := &dto.CreditRequest{}

	json.NewDecoder(r.Body).Decode(request)

	if request.Currency <= 0 {
		err := errors.WithMessage(entities.ErrInvalidParam, "invalid currency")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	balance := entities.NewBalance(userID, entities.Currency(request.Currency))

	err := s.svc.CreditBalance(ctx, balance)
	if err != nil {
		s.log.Error(err)
		s.writeError(w, entities.ErrInternal, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{}{})
}

func (s *Server) writeError(w http.ResponseWriter, err error, code int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := &dto.ErrResponse{
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(resp)
}
