{{ template "boilerplate" }}

package {{ .CRD.Names.Lower }}

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/{{ .ServiceIDClean }}"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/{{ .ServiceIDClean }}/{{ .APIVersion}}"
)

// Setup{{ .CRD.Names.Camel }} adds a controller that reconciles {{ .CRD.Names.Camel }}.
func Setup{{ .CRD.Names.Camel }}(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.{{ .CRD.Names.Camel }}GroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.{{ .CRD.Names.Camel }}{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.{{ .CRD.Names.Camel }}GroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.{{ .CRD.Names.Camel }}) error {
	return nil
}

{{- if .CRD.Ops.ReadOne }}
func (*external) postObserve(_ context.Context, _ *svcapitypes.{{ .CRD.Names.Camel }}, _ *svcsdk.{{ .CRD.Ops.ReadOne.OutputRef.Shape.ShapeName }}, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
{{- else if .CRD.Ops.ReadMany }}
func (*external) postObserve(_ context.Context, _ *svcapitypes.{{ .CRD.Names.Camel }}, _ *svcsdk.{{ .CRD.Ops.ReadMany.OutputRef.Shape.ShapeName }}, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}

func (*external) filterList(_ *svcapitypes.{{ .CRD.Names.Camel }}, list *svcsdk.{{ .CRD.Ops.ReadMany.OutputRef.Shape.ShapeName }}) *svcsdk.{{ .CRD.Ops.ReadMany.OutputRef.Shape.ShapeName }} {
	return list
}
{{ end }}

func (*external) preCreate(context.Context, *svcapitypes.{{ .CRD.Names.Camel }}) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.{{ .CRD.Names.Camel }}, _ *svcsdk.{{ .CRD.Ops.Create.OutputRef.Shape.ShapeName }}, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.{{ .CRD.Names.Camel }}) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.{{ .CRD.Names.Camel }}, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}

{{- if .CRD.Ops.ReadOne }}
func lateInitialize(*svcapitypes.{{ .CRD.Names.Camel }}Parameters,*svcsdk.{{ .CRD.Ops.ReadOne.OutputRef.Shape.ShapeName }}) error {
	return nil
}

func isUpToDate(*svcapitypes.{{ .CRD.Names.Camel }},*svcsdk.{{ .CRD.Ops.ReadOne.OutputRef.Shape.ShapeName }}) bool {
	return true
}

func preGenerate{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }})*svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }} {
	return obj
}

func postGenerate{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }})*svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }} {
	return obj
}
{{- else if .CRD.Ops.ReadMany }}

func lateInitialize(*svcapitypes.{{ .CRD.Names.Camel }}Parameters,*svcsdk.{{ .CRD.Ops.ReadMany.OutputRef.Shape.ShapeName }}) error {
	return nil
}

func isUpToDate(*svcapitypes.{{ .CRD.Names.Camel }},*svcsdk.{{ .CRD.Ops.ReadMany.OutputRef.Shape.ShapeName }}) bool {
	return true
}

func preGenerate{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }})*svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }} {
	return obj
}

func postGenerate{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }})*svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }} {
	return obj
}
{{ end }}

func preGenerate{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}) *svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }} {
	return obj
}

func postGenerate{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}) *svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }} {
	return obj
}

{{- if .CRD.Ops.Delete }}
func preGenerate{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}) *svcsdk.{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }} {
	return obj
}

func postGenerate{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}(_ *svcapitypes.{{ .CRD.Names.Camel }}, obj *svcsdk.{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}) *svcsdk.{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }} {
	return obj
}
{{- end }}
