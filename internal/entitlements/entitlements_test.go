package entitlements

import (
	"testing"

	"zion-english/internal/constants"
)

func TestHasFeatureOnPlanFree(t *testing.T) {
	free := TeacherPlanView{Tier: constants.TeacherPlanTierFree}
	if !hasFeatureOnPlan(free, FeatureAnalyticsTrial) {
		t.Fatal("expected trial analytics on free")
	}
	if hasFeatureOnPlan(free, FeatureExportClasses) {
		t.Fatal("expected export classes gated on free")
	}
}

func TestHasFeatureOnPlanPro(t *testing.T) {
	pro := TeacherPlanView{Tier: constants.TeacherPlanTierPro, IsPro: true}
	if !hasFeatureOnPlan(pro, FeatureExportClasses) {
		t.Fatal("expected export classes on pro")
	}
}
