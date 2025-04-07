//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha2

import (
	"github.com/act3-ai/go-common/pkg/redact"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACEHubInstance) DeepCopyInto(out *ACEHubInstance) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACEHubInstance.
func (in *ACEHubInstance) DeepCopy() *ACEHubInstance {
	if in == nil {
		return nil
	}
	out := new(ACEHubInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BottleSpec) DeepCopyInto(out *BottleSpec) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BottleSpec.
func (in *BottleSpec) DeepCopy() *BottleSpec {
	if in == nil {
		return nil
	}
	out := new(BottleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientConfiguration) DeepCopyInto(out *ClientConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ClientConfigurationSpec.DeepCopyInto(&out.ClientConfigurationSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientConfiguration.
func (in *ClientConfiguration) DeepCopy() *ClientConfiguration {
	if in == nil {
		return nil
	}
	out := new(ClientConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClientConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientConfigurationSpec) DeepCopyInto(out *ClientConfigurationSpec) {
	*out = *in
	if in.Locations != nil {
		in, out := &in.Locations, &out.Locations
		*out = make([]Location, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientConfigurationSpec.
func (in *ClientConfigurationSpec) DeepCopy() *ClientConfigurationSpec {
	if in == nil {
		return nil
	}
	out := new(ClientConfigurationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Database) DeepCopyInto(out *Database) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Database.
func (in *Database) DeepCopy() *Database {
	if in == nil {
		return nil
	}
	out := new(Database)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPU) DeepCopyInto(out *GPU) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPU.
func (in *GPU) DeepCopy() *GPU {
	if in == nil {
		return nil
	}
	out := new(GPU)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HubEnvTemplateSpec) DeepCopyInto(out *HubEnvTemplateSpec) {
	*out = *in
	if in.GPU != nil {
		in, out := &in.GPU, &out.GPU
		*out = new(GPU)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.SharedMemory != nil {
		in, out := &in.SharedMemory, &out.SharedMemory
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.Bottles != nil {
		in, out := &in.Bottles, &out.Bottles
		*out = make([]BottleSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]Port, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HubEnvTemplateSpec.
func (in *HubEnvTemplateSpec) DeepCopy() *HubEnvTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(HubEnvTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Location) DeepCopyInto(out *Location) {
	*out = *in
	out.OAuth = in.OAuth
	if in.Cookies != nil {
		in, out := &in.Cookies, &out.Cookies
		*out = make(map[string]redact.Secret, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Location.
func (in *Location) DeepCopy() *Location {
	if in == nil {
		return nil
	}
	out := new(Location)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthProvider) DeepCopyInto(out *OAuthProvider) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthProvider.
func (in *OAuthProvider) DeepCopy() *OAuthProvider {
	if in == nil {
		return nil
	}
	out := new(OAuthProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Port) DeepCopyInto(out *Port) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Port.
func (in *Port) DeepCopy() *Port {
	if in == nil {
		return nil
	}
	out := new(Port)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerConfiguration) DeepCopyInto(out *ServerConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ServerConfigurationSpec.DeepCopyInto(&out.ServerConfigurationSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerConfiguration.
func (in *ServerConfiguration) DeepCopy() *ServerConfiguration {
	if in == nil {
		return nil
	}
	out := new(ServerConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServerConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerConfigurationSpec) DeepCopyInto(out *ServerConfigurationSpec) {
	*out = *in
	out.DB = in.DB
	in.WebApp.DeepCopyInto(&out.WebApp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerConfigurationSpec.
func (in *ServerConfigurationSpec) DeepCopy() *ServerConfigurationSpec {
	if in == nil {
		return nil
	}
	out := new(ServerConfigurationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ViewerSpec) DeepCopyInto(out *ViewerSpec) {
	*out = *in
	in.ACEHub.DeepCopyInto(&out.ACEHub)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ViewerSpec.
func (in *ViewerSpec) DeepCopy() *ViewerSpec {
	if in == nil {
		return nil
	}
	out := new(ViewerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebApp) DeepCopyInto(out *WebApp) {
	*out = *in
	if in.ACEHubs != nil {
		in, out := &in.ACEHubs, &out.ACEHubs
		*out = make([]ACEHubInstance, len(*in))
		copy(*out, *in)
	}
	if in.Viewers != nil {
		in, out := &in.Viewers, &out.Viewers
		*out = make([]ViewerSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DefaultBottleSelectors != nil {
		in, out := &in.DefaultBottleSelectors, &out.DefaultBottleSelectors
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebApp.
func (in *WebApp) DeepCopy() *WebApp {
	if in == nil {
		return nil
	}
	out := new(WebApp)
	in.DeepCopyInto(out)
	return out
}
