package swagger

import (
	"net/http"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"github.com/swaggest/openapi-go/openapi3"
	v5 "github.com/swaggest/swgui/v5"
)

type RouteMetadata struct {
	Summary     string
	Description string
	Tags        []string
	Input       any
	Output      any
	Params      []requests.Param[string]
}

func NewRouteMetadata() *RouteMetadata {
	return &RouteMetadata{}
}

func (m *RouteMetadata) WithSummary(summary string) *RouteMetadata {
	m.Summary = summary
	return m
}

func (m *RouteMetadata) WithDescription(description string) *RouteMetadata {
	m.Description = description
	return m
}

func (m *RouteMetadata) WithTags(tags ...string) *RouteMetadata {
	m.Tags = append(m.Tags, tags...)
	return m
}

func (m *RouteMetadata) WithInput(input any) *RouteMetadata {
	m.Input = input
	return m
}

func (m *RouteMetadata) WithOutput(output any) *RouteMetadata {
	m.Output = output
	return m
}

func (m *RouteMetadata) WithParam(param requests.Param[string]) *RouteMetadata {
	m.Params = append(m.Params, param)
	return m
}

func (m *RouteMetadata) WithParams(params ...requests.Param[string]) *RouteMetadata {
	m.Params = append(m.Params, params...)
	return m
}

type Registry struct {
	reflector openapi3.Reflector
}

func NewRegistry(title, description, version string) *Registry {
	r := openapi3.Reflector{}
	r.Spec = &openapi3.Spec{
		Openapi: "3.0.3",
		Info: openapi3.Info{
			Title:       title,
			Description: &description,
			Version:     version,
		},
	}
	return &Registry{reflector: r}
}

func (r *Registry) AddRoute(method, path, summary, description string, tags []string, input, output any, params []requests.Param[string]) error {
	op := openapi3.Operation{}
	op.WithSummary(summary)
	op.WithDescription(description)
	op.WithTags(tags...)

	for _, param := range params {
		parameter := openapi3.Parameter{}
		parameter.WithName(param.Name)
		parameter.WithIn(openapi3.ParameterIn(param.In))
		parameter.WithRequired(param.Required)
		if param.Description != "" {
			parameter.WithDescription(param.Description)
		}
		if param.Value != "" {
			parameter.WithExample(param.Value)
		}
		op.WithParameters(parameter.ToParameterOrRef())
	}

	if input != nil {
		if err := r.reflector.SetRequest(&op, input, method); err != nil {
			return err
		}
	}

	if output != nil {
		if err := r.reflector.SetJSONResponse(&op, output, http.StatusOK); err != nil {
			return err
		}
	}

	return r.reflector.Spec.AddOperation(method, path, op)
}

func (r *Registry) GetSpecJSON() ([]byte, error) {
	return r.reflector.Spec.MarshalJSON()
}

func (r *Registry) StartSwaggerUI(addr string) {
	l := logger.GetLogger()
	l.Info("Starting Swagger UI", map[string]any{"address": addr})

	specJSON, err := r.GetSpecJSON()
	if err != nil {
		l.Error("failed to generate openapi spec", err, nil)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(specJSON)
	})

	mux.Handle("/docs/", v5.NewHandler(r.reflector.Spec.Info.Title, "/openapi.json", "/docs/"))

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			l.Error("Swagger UI server failed", err, nil)
		}
	}()
}
