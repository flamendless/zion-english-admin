package entitlements

import (
	"context"
	"strings"

	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type TeacherPlanView struct {
	Tier         constants.TeacherPlanTier
	IsPro        bool
	BillingKind  constants.TeacherPlanBillingKind
	GrantID      int64
	EffectiveEnd string
}

func TeacherPlan(ctx context.Context, q *queries.Queries, teacherID int64) (TeacherPlanView, error) {
	tier, grant, err := database.EffectiveTeacherPlanTier(ctx, q, teacherID)
	if err != nil {
		return TeacherPlanView{}, err
	}
	view := TeacherPlanView{Tier: tier, IsPro: tier == constants.TeacherPlanTierPro}
	if grant != nil {
		view.GrantID = grant.ID
		view.BillingKind = constants.TeacherPlanBillingKind(grant.BillingKind)
		end := utils.NullStringFromAny(grant.EffectiveEnd)
		if end.Valid {
			view.EffectiveEnd = end.String
		}
	}
	return view, nil
}

func HasFeature(ctx context.Context, q *queries.Queries, actor auth.User, role auth.Role, feature Feature) (bool, error) {
	if auth.HasAdminAccess(role) {
		return true, nil
	}
	if role != auth.RoleTeacher {
		return false, nil
	}
	plan, err := TeacherPlan(ctx, q, actor.ID)
	if err != nil {
		return false, err
	}
	return hasFeatureOnPlan(plan, feature), nil
}

func hasFeatureOnPlan(plan TeacherPlanView, feature Feature) bool {
	if plan.IsPro {
		switch feature {
		case FeatureAnalyticsPage, FeatureAnalyticsDateRange, FeatureAnalyticsSummary,
			FeatureAnalyticsTrial, FeatureAnalyticsChartsWeekly, FeatureAnalyticsByStudent,
			FeatureAnalyticsNoShows, FeatureAnalyticsRetention,
			FeatureExportPage, FeatureExportClasses, FeatureExportStudents, FeatureExportSchedule,
			FeatureNotificationsEmail:
			return true
		default:
			return false
		}
	}
	switch feature {
	case FeatureAnalyticsPage, FeatureAnalyticsSummary, FeatureAnalyticsTrial:
		return true
	case FeatureExportPage:
		return true
	default:
		return false
	}
}

func ClampAnalyticsRange(plan TeacherPlanView, requestedStart, requestedEnd string) (string, string, error) {
	monthStart, monthEnd := database.CurrentMonthDateRangePHT()
	if plan.IsPro {
		return requestedStart, requestedEnd, nil
	}
	reqStart := strings.TrimSpace(requestedStart)
	reqEnd := strings.TrimSpace(requestedEnd)
	if reqStart != "" && reqStart < monthStart {
		return "", "", ErrAnalyticsRangeRequiresPro
	}
	if reqEnd != "" && reqEnd > monthEnd {
		return "", "", ErrAnalyticsRangeRequiresPro
	}
	return monthStart, monthEnd, nil
}
