//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package meta

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetaTags) DeepCopyInto(out *MetaTags) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetaTags.
func (in *MetaTags) DeepCopy() *MetaTags {
	if in == nil {
		return nil
	}
	out := new(MetaTags)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceMeta) DeepCopyInto(out *ResourceMeta) {
	*out = *in
	if in.OrchestrationTags != nil {
		in, out := &in.OrchestrationTags, &out.OrchestrationTags
		*out = new(MetaTags)
		(*in).DeepCopyInto(*out)
	}
	if in.ChildTags != nil {
		in, out := &in.ChildTags, &out.ChildTags
		*out = new(MetaTags)
		(*in).DeepCopyInto(*out)
	}
	if in.ServiceTags != nil {
		in, out := &in.ServiceTags, &out.ServiceTags
		*out = new(MetaTags)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceMeta.
func (in *ResourceMeta) DeepCopy() *ResourceMeta {
	if in == nil {
		return nil
	}
	out := new(ResourceMeta)
	in.DeepCopyInto(out)
	return out
}