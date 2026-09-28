package cmd

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

func notificationFromAvatars(ctx context.Context, rows []queries.TblNotification) map[string]frontend.AvatarProps {
	nameToTeacherID := make(map[string]int64)
	for _, row := range rows {
		name := strings.TrimSpace(row.FromName)
		if name == "" {
			continue
		}
		id := sqlNumericToInt64(row.FromTeacherID)
		if id > 0 {
			nameToTeacherID[name] = id
			continue
		}
		if _, ok := nameToTeacherID[name]; !ok {
			nameToTeacherID[name] = 0
		}
	}

	teacherCache := make(map[int64]queries.GetTeacherProfileByIDRow)
	for _, id := range nameToTeacherID {
		if id <= 0 {
			continue
		}
		if _, loaded := teacherCache[id]; loaded {
			continue
		}
		t, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, id)
		if err != nil {
			continue
		}
		teacherCache[id] = t
	}

	avatars := make(map[string]frontend.AvatarProps, len(nameToTeacherID))
	for name, id := range nameToTeacherID {
		if id > 0 {
			if t, ok := teacherCache[id]; ok {
				avatars[name] = buildTeacherListAvatarProps(
					t.ID,
					t.FirstName,
					t.MiddleName,
					t.LastName,
					t.AssignedColor,
					t.ProfilePicture,
				)
				continue
			}
		}
		avatars[name] = notificationNameAvatar(name)
	}
	return avatars
}

func notificationNameAvatar(name string) frontend.AvatarProps {
	color := constants.DefaultTeacherAssignedColor
	if strings.EqualFold(name, "superuser") {
		color = constants.DefaultAssignedColor
	}
	return frontend.AvatarProps{
		Size:          "sm",
		Initials:      utils.PersonInitials("", "", "", name),
		AssignedColor: color,
		HasPicture:    false,
		Alt:           name + " avatar",
	}
}

func notificationFromOptions(rows []queries.TblNotification, avatars map[string]frontend.AvatarProps) []frontend.NotificationFromOption {
	seen := make(map[string]bool)
	names := make([]string, 0)
	for _, row := range rows {
		name := strings.TrimSpace(row.FromName)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)
	opts := make([]frontend.NotificationFromOption, 0, len(names))
	for _, name := range names {
		avatar := avatars[name]
		if avatar.Size == "" {
			avatar = notificationNameAvatar(name)
		}
		opts = append(opts, frontend.NotificationFromOption{
			Value:  name,
			Label:  name,
			Avatar: avatar,
		})
	}
	return opts
}

func notificationItems(rows []queries.TblNotification, avatars map[string]frontend.AvatarProps) []frontend.NotificationItem {
	items := make([]frontend.NotificationItem, 0, len(rows))
	for _, row := range rows {
		fromName := row.FromName
		avatar := avatars[strings.TrimSpace(fromName)]
		if avatar.Size == "" {
			avatar = notificationNameAvatar(fromName)
		}
		items = append(items, frontend.NotificationItem{
			ID:         strconv.FormatInt(row.ID, 10),
			From:       fromName,
			FromAvatar: avatar,
			To:         row.ToName,
			Message:    row.Message,
			CreatedAt:  formatNotificationCreatedAt(row.CreatedAt),
			Read:       row.Read == 1,
		})
	}
	return items
}
