package constants

type TeacherPlanTier string

const (
	TeacherPlanTierFree TeacherPlanTier = "free"
	TeacherPlanTierPro  TeacherPlanTier = "pro"
)

type TeacherPlanBillingKind string

const (
	TeacherPlanBillingMonthly  TeacherPlanBillingKind = "monthly"
	TeacherPlanBillingLifetime TeacherPlanBillingKind = "lifetime"
)

func AllTeacherPlanBillingKinds() []TeacherPlanBillingKind {
	return []TeacherPlanBillingKind{
		TeacherPlanBillingMonthly,
		TeacherPlanBillingLifetime,
	}
}
