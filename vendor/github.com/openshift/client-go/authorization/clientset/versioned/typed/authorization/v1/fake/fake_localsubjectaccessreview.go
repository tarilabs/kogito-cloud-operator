package fake

import (
	v1 "github.com/openshift/api/authorization/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// FakeLocalSubjectAccessReviews implements LocalSubjectAccessReviewInterface
type FakeLocalSubjectAccessReviews struct {
	Fake *FakeAuthorizationV1
	ns   string
}

var localsubjectaccessreviewsResource = schema.GroupVersionResource{Group: "authorization.openshift.io", Version: "v1", Resource: "localsubjectaccessreviews"}

var localsubjectaccessreviewsKind = schema.GroupVersionKind{Group: "authorization.openshift.io", Version: "v1", Kind: "LocalSubjectAccessReview"}

// Create takes the representation of a localSubjectAccessReview and creates it.  Returns the server's representation of the subjectAccessReviewResponse, and an error, if there is any.
func (c *FakeLocalSubjectAccessReviews) Create(localSubjectAccessReview *v1.LocalSubjectAccessReview) (result *v1.SubjectAccessReviewResponse, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(localsubjectaccessreviewsResource, c.ns, localSubjectAccessReview), &v1.SubjectAccessReviewResponse{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.SubjectAccessReviewResponse), err
}
