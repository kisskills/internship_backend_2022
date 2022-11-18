package http

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/flowchartsman/swaggerui"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"service/internal/entities"
	"service/pkg/dto"
	"strconv"
	"time"
)

const (
	userIDURLParam = "user_id"
	stopTimeout    = 5 * time.Second
)

//go:embed doc/swagger.json
var spec []byte

type Server struct {
	router *chi.Mux
	svc    BalanceService
	port   int
	log    *zap.SugaredLogger
	server *http.Server
}

func NewServer(log *zap.SugaredLogger, svc BalanceService, port int) (*Server, error) {
	if log == nil {
		return nil, errors.WithMessage(entities.ErrInvalidParam, "empty logger")
	}

	if svc == nil || svc == BalanceService(nil) {
		return nil, errors.WithMessage(entities.ErrInvalidParam, "empty service")
	}

	if port == 0 {
		return nil, errors.WithMessage(entities.ErrInvalidParam, "empty port")
	}

	router := chi.NewRouter()

	server := &Server{
		router: router,
		svc:    svc,
		port:   port,
		log:    log,
	}

	basePath := "/api/v1"

	router.Route(basePath, func(r chi.Router) {
		r.Get(fmt.Sprintf("/balances/{%s}", userIDURLParam), server.GetUserBalance)
		r.Post(fmt.Sprintf("/balances/{%s}/credit", userIDURLParam), server.CreditBalance)
		r.Post(fmt.Sprintf("/balances/{%s}/reserve", userIDURLParam), server.ReserveFromBalance)
		r.Post(fmt.Sprintf("/balances/{%s}/commit", userIDURLParam), server.CommitReserve)
		r.Get(fmt.Sprintf("/balances/{%s}/operations", userIDURLParam), server.ListOperations)
	})

	router.Mount("/swagger/", server.SwaggerHandler(spec))

	return server, nil
}

func (s *Server) Run(ctx context.Context) {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	s.log.Infof("server listen %s", addr)

	go s.stopProcess(ctx)

	s.server = &http.Server{Addr: addr, Handler: s.router}

	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.log.Fatal(err)
	}
}

func (s *Server) stopProcess(ctx context.Context) {
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Error(err)
	}
}

func (s *Server) SwaggerHandler(spec []byte) http.Handler {
	return http.StripPrefix("/swagger", swaggerui.Handler(spec))
}

// GetUserBalance returns user balance info
// swagger:operation GET /balances/{user_id} public GetUserBalance
//
// # GetUserBalance returns user balance info
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: user_id
//     in: path
//     description: "User id"
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	 description: Success response
//	 schema:
//	  "$ref": "#/definitions/Balance"
//	'400':
//	 description: Bad response
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'404':
//	 description: Not found
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'500':
//	 description: Internal error
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
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
		Currency: int(balance.Value()),
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(response); err != nil {
		s.log.Info(err)
	}
}

// CreditBalance credit value to user balance
// swagger:operation POST /balances/{user_id}/credit public CreditBalance
//
// # CreditBalance credit value to user balance
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: user_id
//     in: path
//     description: "User id"
//     required: true
//     type: string
//   - name: body
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/CreditRequest'
//
// responses:
//
//	'200':
//	 description: Success response
//	'400':
//	 description: Bad response
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'404':
//	 description: Not found
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'500':
//	 description: Internal error
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
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

	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		err = errors.WithMessage(entities.ErrInvalidParam, "encode body")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	if request.Currency <= 0 {
		err := errors.WithMessage(entities.ErrInvalidParam, "invalid currency")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	err = s.svc.CreditBalance(ctx, userID, entities.Currency(request.Currency))
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

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(struct{}{}); err != nil {
		s.log.Info(err)
	}
}

// ReserveFromBalance reserve value from user's balance
// swagger:operation POST /balances/{user_id}/reserve public ReserveFromBalance
//
// # ReserveFromBalance reserve value from user's balance
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: user_id
//     in: path
//     description: "User id"
//     required: true
//     type: string
//   - name: body
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/ReserveRequest'
//
// responses:
//
//	'200':
//	 description: Success response
//	'400':
//	 description: Bad response
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'404':
//	 description: Not found
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'409':
//	 description: Conflict
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'500':
//	 description: Internal error
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
func (s *Server) ReserveFromBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := chi.URLParam(r, userIDURLParam)
	if userID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty operation id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	request := &dto.ReserveRequest{}

	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		err = errors.WithMessage(entities.ErrInvalidParam, "encode body")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	if request.Currency <= 0 {
		err := errors.WithMessage(entities.ErrInvalidParam, "invalid currency")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	if request.ServiceID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty service id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	if request.OrderID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty order id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	err = s.svc.ReserveFromBalance(ctx, userID, request.ServiceID, request.OrderID, entities.Currency(request.Currency))
	if errors.Is(err, entities.ErrNotFound) {
		s.log.Error(err)
		s.writeError(w, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, entities.ErrReserveAlreadyExists) {
		s.log.Error(err)
		s.writeError(w, err, http.StatusConflict)
		return
	}
	if errors.Is(err, entities.ErrReserveInvalidValue) {
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}
	if err != nil {
		s.log.Error(err)
		s.writeError(w, entities.ErrInternal, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(struct{}{}); err != nil {
		s.log.Info(err)
	}
}

// CommitReserve commit reserve
// swagger:operation POST /balances/{user_id}/commit public CommitReserve
//
// # CommitReserve commit reserve
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: user_id
//     in: path
//     description: "User id"
//     required: true
//     type: string
//   - name: body
//     in: body
//     required: true
//     schema:
//     $ref: '#/definitions/CommitReserveRequest'
//
// responses:
//
//	'200':
//	 description: Success response
//	'400':
//	 description: Bad response
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'404':
//	 description: Not found
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'409':
//	 description: Conflict
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'500':
//	 description: Internal error
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
func (s *Server) CommitReserve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := chi.URLParam(r, userIDURLParam)
	if userID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty operation id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	request := &dto.CommitReserveRequest{}

	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		err = errors.WithMessage(entities.ErrInvalidParam, "encode body")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	if request.Currency <= 0 {
		err := errors.WithMessage(entities.ErrInvalidParam, "invalid currency")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	if request.ServiceID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty service id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	if request.OrderID == "" {
		err := errors.WithMessage(entities.ErrInvalidParam, "empty order id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	err = s.svc.CommitReserve(ctx, userID, request.ServiceID, request.OrderID, entities.Currency(request.Currency))
	if errors.Is(err, entities.ErrInvalidParam) {
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}
	if errors.Is(err, entities.ErrNotFound) {
		s.log.Error(err)
		s.writeError(w, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, entities.ErrCommitInvalidValue) {
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}
	if err != nil {
		s.log.Error(err)
		s.writeError(w, entities.ErrInternal, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(struct{}{}); err != nil {
		s.log.Info(err)
	}
}

// ListOperations list balance operations
// swagger:operation GET /balances/{user_id}/operations public ListOperations
//
// # ListOperations list balance operations
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: user_id
//     in: path
//     description: "User id"
//     required: true
//     type: string
//   - name: limit
//     in: query
//     description: "Limit"
//     required: false
//     type: integer
//   - name: offset
//     in: query
//     description: "Offset"
//     required: false
//     type: integer
//   - name: order_by
//     in: query
//     description: "Field order by. date or value"
//     required: false
//     type: string
//   - name: desc
//     in: query
//     description: "Response in desc order"
//     required: false
//     type: boolean
//
// responses:
//
//	'200':
//	 description: Success response
//	'400':
//	 description: Bad response
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'404':
//	 description: Not found
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'409':
//	 description: Conflict
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
//	'500':
//	 description: Internal error
//	 schema:
//	  "$ref": "#/definitions/ErrResponse"
func (s *Server) ListOperations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	userID := chi.URLParam(r, userIDURLParam)
	if userID == "" {
		err = errors.WithMessage(entities.ErrInvalidParam, "empty operation id")
		s.log.Error(err)
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	var offset int
	offsetParam := r.URL.Query().Get("offset")
	if offsetParam != "" {
		offset, err = strconv.Atoi(offsetParam)
		if err != nil {
			err = errors.WithMessage(entities.ErrInvalidParam, "invalid offset param")
			s.log.Error(err)
			s.writeError(w, err, http.StatusBadRequest)
			return
		}
	}

	limit := 10
	limitParam := r.URL.Query().Get("limit")
	if limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			err = errors.WithMessage(entities.ErrInvalidParam, "invalid limit param")
			s.log.Error(err)
			s.writeError(w, err, http.StatusBadRequest)
			return
		}
	}

	orderBy := entities.Date
	orderByParam := r.URL.Query().Get("order_by")
	if orderByParam != "" {
		if orderByParam != entities.Date && orderByParam != entities.Value {
			err = errors.WithMessage(entities.ErrInvalidParam, "invalid order by param")
			s.log.Error(err)
			s.writeError(w, err, http.StatusBadRequest)
			return
		}
		orderBy = orderByParam
	}

	desc := true
	descParam := r.URL.Query().Get("desc")
	if descParam != "" {
		desc, err = strconv.ParseBool(descParam)
		if err != nil {
			err = errors.WithMessage(entities.ErrInvalidParam, "invalid desc by param")
			s.log.Error(err)
			s.writeError(w, err, http.StatusBadRequest)
			return
		}
	}

	operations, err := s.svc.ListOperations(ctx, userID, limit, offset, orderBy, desc)
	if err != nil {
		s.log.Error(err)
		s.writeError(w, entities.ErrInternal, http.StatusInternalServerError)
		return
	}

	operationDtos := make([]dto.Operation, 0, len(operations))

	for _, operation := range operations {
		operationDtos = append(operationDtos, dto.ToOperation(operation))
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(operationDtos); err != nil {
		s.log.Info(err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, err error, code int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := &dto.ErrResponse{
		Message: err.Error(),
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		s.log.Info(err)
	}
}
