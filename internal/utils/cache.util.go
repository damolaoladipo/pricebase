package utils

type cacheKeys struct {
	Roles       string
	Users       string
	SessionJTI  string
	UserSession string
	Permission  string
}

var CacheKeys = cacheKeys{
	Roles:       "pricebase.auth.cached.roles",
	Users:       "pricebase.auth.cached.users",
	SessionJTI:  "pricebase.cache.jti",
	UserSession: "pricebase.cache.user.session",
	Permission:  "pricebase.cache.permission",
}

func ComputeKey(env, key string) string {
	if env == "production" {
		return key + ".prod"
	}

	return key
}
