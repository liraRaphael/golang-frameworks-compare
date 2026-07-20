package chi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

type adapter struct {
	router chi.Router
	handle ports.Handler
	tracer tracing.Tracer
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
	return &adapter{
		router: chi.NewRouter(),
		handle: handle,
		tracer: tracing.NewTracer("chi-adapter"),
	}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller) {
	a.router.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.tracer.Start(r.Context(), "http.request")
		defer span.End()

		l := logger.FromContext(ctx)
		ctx = logger.ToContext(ctx, l)

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   r.Method,
			"path":     r.URL.Path,
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				resp := a.handle.ResolveError(ctx, err)
				a.writeResponse(w, resp)
				return
			}
		}

		headers := domain.NewHttpParam()
		headers.SetAll(r.Header)

		query := domain.NewHttpParam()
		query.SetAll(r.URL.Query())

		pathParams := domain.NewHttpParam()
		rctx := chi.RouteContext(r.Context())
		if rctx != nil {
			for i, key := range rctx.URLParams.Keys {
				value := rctx.URLParams.Values[i]
				pathParams.Set(key, value)
			}
		}

		req := requests.NewRequestFromParams(body, headers, query, pathParams, &requests.HelloRequest{})
		result, err := ctrl.WrapperExecute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			a.writeResponse(w, resp)
			return
		}

		resp := a.handle.Handle(ctx, http.StatusOK, result, nil)
		a.writeResponse(w, resp)
	})
}

func (a *adapter) writeResponse(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(resp)
	w.Write(b)
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	return http.ListenAndServe(addr, a.router)
}
