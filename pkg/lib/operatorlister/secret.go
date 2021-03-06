package operatorlister

import (
	"sync"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	corev1 "k8s.io/client-go/listers/core/v1"
)

type UnionSecretLister struct {
	secretListers map[string]corev1.SecretLister
	secretLock    sync.RWMutex
}

// List lists all Secrets in the indexer.
func (usl *UnionSecretLister) List(selector labels.Selector) (ret []*v1.Secret, err error) {
	usl.secretLock.RLock()
	defer usl.secretLock.RUnlock()

	var set map[types.UID]*v1.Secret
	for _, sl := range usl.secretListers {
		secrets, err := sl.List(selector)
		if err != nil {
			return nil, err
		}

		for _, secret := range secrets {
			set[secret.GetUID()] = secret
		}
	}

	for _, secret := range set {
		ret = append(ret, secret)
	}

	return
}

// Secrets returns an object that can list and get Secrets.
func (usl *UnionSecretLister) Secrets(namespace string) corev1.SecretNamespaceLister {
	usl.secretLock.RLock()
	defer usl.secretLock.RUnlock()

	// Check for specific namespace listers
	if sl, ok := usl.secretListers[namespace]; ok {
		return sl.Secrets(namespace)
	}

	// Check for any namespace-all listers
	if sl, ok := usl.secretListers[metav1.NamespaceAll]; ok {
		return sl.Secrets(namespace)
	}

	return nil
}

func (usl *UnionSecretLister) RegisterSecretLister(namespace string, lister corev1.SecretLister) {
	usl.secretLock.Lock()
	defer usl.secretLock.Unlock()

	if usl.secretListers == nil {
		usl.secretListers = make(map[string]corev1.SecretLister)
	}
	usl.secretListers[namespace] = lister
}

func (l *coreV1Lister) RegisterSecretLister(namespace string, lister corev1.SecretLister) {
	l.secretLister.RegisterSecretLister(namespace, lister)
}

func (l *coreV1Lister) SecretLister() corev1.SecretLister {
	return l.secretLister
}
