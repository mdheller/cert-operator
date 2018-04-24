package vaultpki

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/deletionallowedcontext"

	"github.com/giantswarm/cert-operator/service/controller/v2/key"
)

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	vaultPKIStateToDelete, err := toVaultPKIState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if vaultPKIStateToDelete.Backend != nil || vaultPKIStateToDelete.CACertificate != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the Vault PKI in the Vault API")

		err := r.vaultPKI.DeleteBackend(key.ClusterID(customObject))
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the Vault PKI in the Vault API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the Vault PKI does not need to be deleted from the Vault API")
	}

	return nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentVaultPKIState, err := toVaultPKIState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredVaultPKIState, err := toVaultPKIState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var vaultPKIStateToDelete VaultPKIState
	if deletionallowedcontext.IsDeletionAllowed(ctx) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the Vault PKI has to be deleted")

		if currentVaultPKIState.Backend == nil {
			vaultPKIStateToDelete.Backend = desiredVaultPKIState.Backend
		}
		if currentVaultPKIState.CACertificate == "" {
			vaultPKIStateToDelete.CACertificate = desiredVaultPKIState.CACertificate
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found out if the Vault PKI has to be deleted")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing delete state because Vault PKIs are not allowed to be deleted")
	}

	return vaultPKIStateToDelete, nil
}