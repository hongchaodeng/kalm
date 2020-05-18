/*

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

package controllers

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	cmmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"math/big"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	cmv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	corev1alpha1 "github.com/kapp-staging/kapp/controller/api/v1alpha1"
)

// HttpsCertIssuerReconciler reconciles a HttpsCertIssuer object
type HttpsCertIssuerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.kapp.dev,resources=httpscertissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kapp.dev,resources=httpscertissuers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete

func (r *HttpsCertIssuerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("httpscertissuer", req.NamespacedName)

	// your logic here
	var httpsCertIssuer corev1alpha1.HttpsCertIssuer
	if err := r.Get(ctx, req.NamespacedName, &httpsCertIssuer); err != nil {
		err = client.IgnoreNotFound(err)
		if err != nil {
			log.Error(err, "fail to get HttpsCertIssuer")
		}

		return ctrl.Result{}, err
	}

	if httpsCertIssuer.Spec.CAForTest != nil {
		return r.ReconcileCAForTest(ctx, httpsCertIssuer)
	}

	if httpsCertIssuer.Spec.ACMECloudFlare != nil {
		return r.ReconcileACMECloudFlare(ctx, httpsCertIssuer)
	}

	return ctrl.Result{}, nil
}

func (r *HttpsCertIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.HttpsCertIssuer{}).
		Owns(&cmv1alpha2.Issuer{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *HttpsCertIssuerReconciler) ReconcileCAForTest(ctx context.Context, certIssuer corev1alpha1.HttpsCertIssuer) (ctrl.Result, error) {
	caSecretName := certIssuer.Name

	// auto generate tls secret for our CA
	sec := corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: certManagerNamespace,
		Name:      caSecretName,
	}, &sec); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		key, crt, err := r.generateRandomPrvKeyAndCrtForCA()
		if err != nil {
			return ctrl.Result{}, err
		}

		sec := corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Namespace: certManagerNamespace,
				Name:      caSecretName,
			},
			StringData: map[string]string{
				"tls.key": string(key),
				"tls.crt": string(crt),
			},
			Type: "kubernetes.io/tls",
		}

		if err := ctrl.SetControllerReference(&certIssuer, &sec, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, &sec); err != nil {
			return ctrl.Result{}, err
		}

		r.Log.Info("secret created")
	}

	expectedClusterIssuer := cmv1alpha2.ClusterIssuer{
		ObjectMeta: v1.ObjectMeta{
			Name: certIssuer.Name,
		},
		Spec: cmv1alpha2.IssuerSpec{
			IssuerConfig: cmv1alpha2.IssuerConfig{
				CA: &cmv1alpha2.CAIssuer{
					SecretName: caSecretName,
				},
			},
		},
	}

	// start our CA using secret
	clusterIssuer := cmv1alpha2.ClusterIssuer{}

	err := r.Get(ctx, types.NamespacedName{
		Name: certIssuer.Name,
	}, &clusterIssuer)

	if err != nil {
		if errors.IsNotFound(err) {
			clusterIssuer = expectedClusterIssuer

			if err := ctrl.SetControllerReference(&certIssuer, &expectedClusterIssuer, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}

			r.Log.Info("creating clusterIssuer")
			if err := r.Create(ctx, &expectedClusterIssuer); err != nil {
				r.Log.Error(err, "fail create clusterIssuer")
				return ctrl.Result{}, err
			}
		} else {
			clusterIssuer.Spec = expectedClusterIssuer.Spec

			r.Log.Info("updating clusterIssuer")
			if err := r.Update(ctx, &expectedClusterIssuer); err != nil {
				r.Log.Error(err, "fail update clusterIssuer")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

const certManagerNamespace = "cert-manager"

// config ACME cloudflare
func (r *HttpsCertIssuerReconciler) ReconcileACMECloudFlare(ctx context.Context, certIssuer corev1alpha1.HttpsCertIssuer) (ctrl.Result, error) {

	acmeSpec := certIssuer.Spec.ACMECloudFlare
	email := acmeSpec.Email

	apiTokenSecretName := acmeSpec.APITokenSecretName
	// for clusterIssuer, secret has to be in ns: cert-manager to be found by cert-mgr
	apiTokenSecretNamespace := certManagerNamespace

	issuerName := certIssuer.Name
	//curNs := certIssuer.Namespace

	apiTokenSecret := corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: apiTokenSecretNamespace,
		Name:      apiTokenSecretName,
	}, &apiTokenSecret); err != nil {
		r.Log.Error(err, fmt.Sprintf("fail to get secret %s", apiTokenSecretName))

		if certIssuer.Status.OK {
			certIssuer.Status.OK = false
			r.Status().Update(ctx, &certIssuer)
		}

		return ctrl.Result{}, err
	}

	var secKey string
	for k, v := range apiTokenSecret.Data {
		if v == nil {
			continue
		}

		secKey = k
	}

	if secKey == "" {
		err := fmt.Errorf("secret %s has no key", apiTokenSecretName)
		r.Log.Error(err, "")

		if certIssuer.Status.OK {
			certIssuer.Status.OK = false
			r.Status().Update(ctx, &certIssuer)
		}

		return ctrl.Result{}, err
	}

	// ref: https://cert-manager.io/docs/configuration/acme/dns01/cloudflare/
	expectedClusterIssuer := cmv1alpha2.ClusterIssuer{
		ObjectMeta: v1.ObjectMeta{
			Name: issuerName,
		},
		Spec: cmv1alpha2.IssuerSpec{
			IssuerConfig: cmv1alpha2.IssuerConfig{
				ACME: &v1alpha2.ACMEIssuer{
					Email: email,
					//Server: "https://acme-staging-v02.api.letsencrypt.org/directory",
					Server: "https://acme-v02.api.letsencrypt.org/directory",
					PrivateKey: cmmetav1.SecretKeySelector{ // what is this prvKey used for?
						LocalObjectReference: cmmetav1.LocalObjectReference{
							Name: getPrvKeyNameForIssuer(certIssuer),
						},
					},
					Solvers: []v1alpha2.ACMEChallengeSolver{
						{
							DNS01: &v1alpha2.ACMEChallengeSolverDNS01{
								Cloudflare: &v1alpha2.ACMEIssuerDNS01ProviderCloudflare{
									Email: email,
									APIToken: &cmmetav1.SecretKeySelector{
										LocalObjectReference: cmmetav1.LocalObjectReference{
											Name: apiTokenSecretName,
										},
										Key: secKey,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	clusterIssuer := cmv1alpha2.ClusterIssuer{}
	var isNew bool
	if err := r.Get(ctx, types.NamespacedName{
		Name: issuerName,
	}, &clusterIssuer); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		clusterIssuer = expectedClusterIssuer
		isNew = true
	}

	if isNew {
		if err := ctrl.SetControllerReference(&certIssuer, &clusterIssuer, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		r.Log.Info("creating clusterIssuer")
		if err := r.Create(ctx, &clusterIssuer); err != nil {
			r.Log.Error(err, "fail create clusterIssuer")
			return ctrl.Result{}, err
		}
	} else {
		clusterIssuer.Spec = expectedClusterIssuer.Spec

		if err := r.Update(ctx, &clusterIssuer); err != nil {
			r.Log.Error(err, "fail update clusterIssuer")
			return ctrl.Result{}, err
		}
	}

	certIssuer.Status.OK = true
	if err := r.Status().Update(ctx, &certIssuer); err != nil {
		return ctrl.Result{}, err
	}

	//todo
	//conditions := clusterIssuer.Status.Conditions
	//latestCondition := conditions[len(conditions) - 1]
	//latestCondition.Type
	//if clusterIssuer.Status.OK !=

	return ctrl.Result{}, nil
}

func getPrvKeyNameForIssuer(issuer corev1alpha1.HttpsCertIssuer) string {
	return fmt.Sprintf("kapp-prvkey-%s", issuer.Name)
}

func (r *HttpsCertIssuerReconciler) generateRandomPrvKeyAndCrtForCA() (prvKey []byte, crt []byte, err error) {
	//priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	//if err != nil {
	//	log.Fatalf("Failed to generate private key: %v", err)
	//	return nil, nil, err
	//}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"Kapp CA for Test Co"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, caPrivKey.Public(), caPrivKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
		return nil, nil, err
	}

	var certOutBuf bytes.Buffer
	if err := pem.Encode(&certOutBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to cert.pem: %v", err)
		return nil, nil, err
	}

	var keyOutBuf bytes.Buffer
	privBytes := x509.MarshalPKCS1PrivateKey(caPrivKey)
	if err := pem.Encode(&keyOutBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to key.pem: %v", err)
		return nil, nil, err
	}

	return keyOutBuf.Bytes(), certOutBuf.Bytes(), nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}
