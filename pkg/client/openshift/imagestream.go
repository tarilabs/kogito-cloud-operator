package openshift

import (
	"encoding/json"
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"

	dockerv10 "github.com/openshift/api/image/docker10"
	imgv1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	//ImageTagLatest is the default name for the latest image tag
	ImageTagLatest = "latest"
	// ImageLabelForExposeServices is the label defined in images to identify ports that need to be exposed by the container
	ImageLabelForExposeServices = "io.openshift.expose-services"
)

// ImageStreamInterface exposes OpenShift ImageStream operations
type ImageStreamInterface interface {
	FetchDockerImage(key types.NamespacedName) (*dockerv10.DockerImage, error)
	FecthTag(key types.NamespacedName, tag string) (*imgv1.ImageStreamTag, error)
	CreateTagIfNotExists(image *imgv1.ImageStreamTag) (bool, error)
}

func newImageStream(c *client.Client) ImageStreamInterface {
	client.MustEnsureClient(c)
	return &imageStream{
		client: c,
	}
}

type imageStream struct {
	client *client.Client
}

// FecthTag fetches for a particular imagestreamtag on OpenShift cluster.
// If tag is nil or empty, will search for "latest".
// Returns nil if the object was not found.
func (i *imageStream) FecthTag(key types.NamespacedName, tag string) (*imgv1.ImageStreamTag, error) {
	if len(tag) == 0 {
		tag = ImageTagLatest
	}
	tagRefName := fmt.Sprintf("%s:%s", key.Name, tag)
	isTag, err := i.client.ImageCli.ImageStreamTags(key.Namespace).Get(tagRefName, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		log.Debugf("Image '%s' not found", tagRefName)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return isTag, err
}

// FetchDockerImage fetches a docker image based on a ImageStreamTag with the defined key (namespace and name).
// Returns nil if not found
func (i *imageStream) FetchDockerImage(key types.NamespacedName) (*dockerv10.DockerImage, error) {
	dockerImage := &dockerv10.DockerImage{}
	isTag, err := i.FecthTag(key, "")
	if err != nil {
		return nil, err
	} else if isTag == nil {
		return nil, nil
	}
	log.Debugf("Found image '%s' in the namespace '%s'", key.Name, key.Namespace)
	// is there any metadata to read from?
	if len(isTag.Image.DockerImageMetadata.Raw) != 0 {
		err = json.Unmarshal(isTag.Image.DockerImageMetadata.Raw, dockerImage)
		if err != nil {
			return nil, err
		}
		return dockerImage, nil
	}

	log.Warnf("Can't find any metadata in the docker image for the imagestream '%s' in the namespace '%s'", key.Name, key.Namespace)
	return nil, nil
}

// CreateTagIfNotExists will create a new ImageStreamTag if not exists
func (i *imageStream) CreateTagIfNotExists(is *imgv1.ImageStreamTag) (bool, error) {
	is, err := i.client.ImageCli.ImageStreamTags(is.Namespace).Create(is)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Debugf("Error while creating Image Stream Tag '%s' in namespace '%s'", is.Name, is.Namespace)
		return false, err
	} else if errors.IsAlreadyExists(err) {
		log.Debug("Image Stream Tag already exists in the namespace")
		return false, nil
	}
	log.Debugf("Image Stream Tag %s created in namespace %s", is.Name, is.Namespace)
	return true, nil
}
