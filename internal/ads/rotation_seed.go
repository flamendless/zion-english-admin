package ads

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"time"

	"zion-english/internal/conf"
	"zion-english/internal/constants"
)

const viewerSeedCookieName = "ze_ad_viewer"

func viewerSeedFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	c, err := r.Cookie(viewerSeedCookieName)
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}

func ensureViewerSeedCookie(w http.ResponseWriter, r *http.Request) {
	if w == nil || r == nil {
		return
	}
	if seed := viewerSeedFromRequest(r); seed != "" {
		return
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return
	}
	val := base64.RawURLEncoding.EncodeToString(b[:])
	cookie := &http.Cookie{
		Name:     viewerSeedCookieName,
		Value:    val,
		Path:     rotationCookiePath(),
		MaxAge:   86400 * 365,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if conf.Conf().IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func rotationSlotKey(ad CatalogAd, zone constants.AdZone, r *http.Request) string {
	switch ad.RandomizeKind {
	case constants.AdRandomizePerPage:
		return fmt.Sprintf("page:%d:%s:%s", ad.ID, zone, requestPagePath(r))
	case constants.AdRandomizeTimer:
		bucket := time.Now().Unix() / int64(timerMaxAge(ad.TimerInterval))
		return fmt.Sprintf("timer:%d:%s:%d", ad.ID, zone, bucket)
	default:
		return fmt.Sprintf("session:%d:%s", ad.ID, zone)
	}
}

func deterministicProductInUnused(seed, slotKey string, unused []ProductOption) (int64, bool) {
	if seed == "" || len(unused) == 0 {
		return 0, false
	}
	sum := sha256.Sum256([]byte(seed + "\x00" + slotKey))
	idx := int(binary.BigEndian.Uint64(sum[:8]) % uint64(len(unused)))
	return unused[idx].ProductID, true
}
