package controllers

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ComponentReconcilerTask) getNameForPermission() string {
	name := fmt.Sprintf("kalm-permission-%s", r.component.Name)
	return name
}

func (r *ComponentReconcilerTask) reconcilePermission() error {
	if r.component == nil || r.component.Spec.RunnerPermission == nil {
		return nil
	}

	permission := r.component.Spec.RunnerPermission
	name := r.getNameForPermission()

	// serviceAccount
	var sa corev1.ServiceAccount
	err := r.Get(
		r.ctx,
		types.NamespacedName{Name: name, Namespace: r.component.Namespace},
		&sa)

	if errors.IsNotFound(err) {
		err := r.Create(r.ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: r.component.Namespace,
			},
		})

		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if permission.RoleType == "clusterRole" {
		// clusterRole
		desiredClusterRole := rbacV1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Rules:      permission.Rules,
		}

		var cr rbacV1.ClusterRole
		err := r.Get(r.ctx, types.NamespacedName{Name: name}, &cr)
		if errors.IsNotFound(err) {
			err := r.Create(r.ctx, &desiredClusterRole)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// ensure
		}

		//binding
		desiredCRB := rbacV1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			RoleRef: rbacV1.RoleRef{
				APIGroup: rbacV1.GroupName,
				Kind:     "ClusterRole",
				Name:     name,
			},
			Subjects: []rbacV1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      name,
					Namespace: r.component.Namespace,
				},
			},
		}

		var crb rbacV1.ClusterRoleBinding
		err = r.Get(r.ctx, types.NamespacedName{Name: name}, &crb)
		if errors.IsNotFound(err) {
			if err := r.Create(r.ctx, &desiredCRB); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			//todo ensure
		}
	} else {
		// role
		desiredRole := rbacV1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: r.component.Namespace},
			Rules:      permission.Rules,
		}

		var cr rbacV1.Role
		err := r.Get(r.ctx, types.NamespacedName{Name: name, Namespace: r.component.Namespace}, &cr)
		if errors.IsNotFound(err) {
			err := r.Create(r.ctx, &desiredRole)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// ensure
		}

		//binding
		desiredRB := rbacV1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: r.component.Namespace,
			},
			RoleRef: rbacV1.RoleRef{
				APIGroup: rbacV1.GroupName,
				Kind:     "Role",
				Name:     name,
			},
			Subjects: []rbacV1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      name,
					Namespace: r.component.Namespace,
				},
			},
		}

		var rb rbacV1.RoleBinding
		err = r.Get(r.ctx, types.NamespacedName{Name: name, Namespace: r.component.Namespace}, &rb)
		if errors.IsNotFound(err) {
			if err := r.Create(r.ctx, &desiredRB); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			//todo ensure
		}
	}

	return nil
}
