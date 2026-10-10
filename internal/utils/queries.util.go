package utils

// Queries for the user table

// Rules: Never write
// This is a bad practice because it is not safe and can lead to SQL injection.

// queries := "INSERT INTO users (email) VALUES ('" + user.Email + "')"
// query := "INSERT INTO users (email) VALUES ('" + user.Email + "')"
// query := "SELECT * FROM users WHERE email = '" + user.Email + "'"
// query := "UPDATE users SET email = '" + user.Email + "' WHERE id = '" + user.ID + "'"
// query := "DELETE FROM users WHERE id = '" + user.ID + "'"

// Rules apply to all SQL operations, including:
//
// - INSERT
// - SELECT
// - UPDATE
// - DELETE
// - WHERE
// - JOIN
// - HAVING
// - LIMIT / OFFSET when values are dynamic
// - INSERT ... RETURNING
// - UPDATE ... RETURNING
// - DELETE ... RETURNING
// - Subqueries

var USER_CREATE = `
    INSERT INTO users (
        id,
        user_type,
        first_name,
        last_name,
        email,
        password,
        saved_password,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id
    )
    VALUES (
        $1, $2, $3, $4, $5, $6,
        $7, $8, $9, $10, $11,
        $12, $13, $14, $15,
        $16, $17, $18, $19, $20,
        $21, $22, $23, $24, $25,
        $26, $27, $28, $29, $30,
        $31, $32, $33, $34, $35,
        $36
    )
    RETURNING
        id,
        email,
        password,
        first_name,
        last_name,
        saved_password,
        user_type,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        created_at,
        updated_at,
        version
`

var USER_CHECK_BY_EMAIL = `
    SELECT EXISTS (
        SELECT 1
        FROM users
        WHERE email = $1
    );
`

var USER_FIND_BY_EMAIL = `
    SELECT
        id,
        first_name,
        last_name,
        email,
        saved_password,
        user_type,
        role_ids,
        permissions,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        created_at,
        updated_at
    FROM users
    WHERE email = $1;
`

var USER_FIND_BY_PHONE = `
    SELECT
        id,
        first_name,
        last_name,
        email,
        saved_password,
        user_type,
        role_ids,
        permissions,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        created_at,
        updated_at
    FROM users
    WHERE phone_number = $1 OR alt_phone = $1
    LIMIT 1;
`

var USER_FIND_BY_OTP = `
    SELECT
        id,
        first_name,
        last_name,
        email,
        saved_password,
        user_type,
        role_ids,
        permissions,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        otp_code,
        otp_code_expire,
        otp_type,
        created_at,
        updated_at
    FROM users
    WHERE otp_code = $1 AND otp_code_expire > NOW()
    LIMIT 1;
`

var USER_FIND_BY_EMAIL_OTP = `
    SELECT
        id,
        first_name,
        last_name,
        email,
        saved_password,
        user_type,
        role_ids,
        permissions,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        otp_code,
        otp_code_expire,
        otp_type,
        created_at,
        updated_at
    FROM users
    WHERE email = $1 AND otp_code = $2 AND otp_type = $3
    LIMIT 1;
`

var USER_UPDATE_USER_TYPE = `
    UPDATE users
    SET user_type = $2,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_LOGIN = `
    UPDATE users
    SET login = $2,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_ACCOUNT_STATUS = `
    UPDATE users
    SET is_activated = $2,
        is_active = $3,
        is_locked = $4,
        login_limit = $5,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_OTP = `
    UPDATE users
    SET otp_code = $2,
        otp_code_expire = $3,
        otp_type = $4,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_FIND_BY_ID = `
    SELECT
        id,
        first_name,
        last_name,
        email,
        saved_password,
        user_type,
        role_ids,
        permissions,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        created_at,
        updated_at
    FROM users
    WHERE id = $1;
`

var USER_UPDATE_SAVED_PASSWORD = `
    UPDATE users
    SET saved_password = $2,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_API_KEYS = `
    UPDATE users
    SET api_key = $2,
        api_keys = $3,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_ROLES_PERMISSIONS = `
    UPDATE users
    SET role_ids = $2,
        permissions = $3,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_LOGIN_LIMIT = `
    UPDATE users
    SET login_limit = $2,
        is_locked = $3,
        locked_until = $4,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_RESET_TOKEN = `
    UPDATE users
    SET reset_password_token = $2,
        reset_password_token_expire = $3,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_ACTIVATION_TOKEN = `
    UPDATE users
    SET activation_token = $2,
        activation_token_expire = $3,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_UPDATE_INVITE_TOKEN = `
    UPDATE users
    SET invite_token = $2,
        invite_token_expire = $3,
        updated_at = NOW()
    WHERE id = $1;
`

var USER_FIND_BY_SECRET_API_KEY = `
    SELECT
        id,
        first_name,
        last_name,
        email,
        saved_password,
        user_type,
        api_key,
        api_keys,
        role_ids,
        permissions,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        created_at,
        updated_at
    FROM users
    WHERE api_key->>'secret' = $1
      AND api_key->>'secretToken' = $2
    LIMIT 1;
`

var USER_FIND_BY_PUBLIC_API_KEY = `
    SELECT
        id,
        first_name,
        last_name,
        email,
        saved_password,
        user_type,
        api_key,
        api_keys,
        role_ids,
        permissions,
        is_super,
        is_admin,
        is_business,
        is_user,
        is_team,
        is_activated,
        is_active,
        is_welcomed,
        is_locked,
        login_limit,
        locked_until,
        business_name,
        business_id,
        login,
        onboard,
        register,
        status,
        oauth,
        country,
        notifications_id,
        devices_id,
        middle_name,
        avatar,
        logo,
        phone_number,
        phone_code,
        country_phone,
        alt_phone,
        password_type,
        passcode,
        team_role_id,
        settings_id,
        created_at,
        updated_at
    FROM users
    WHERE api_key->>'public' = $1
      AND api_key->>'publicToken' = $2
    LIMIT 1;
`

var ROLE_INSERT = `
    INSERT INTO roles (
        id,
        name,
        description,
        slug,
        permissions
    )
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (name) DO NOTHING
`

var ROLE_FIND_BY_ID = `
    SELECT
        id,
        name,
        description,
        slug,
        permissions,
        created_at,
        updated_at
    FROM roles
    WHERE id = $1;
`

var ROLE_FIND_BY_NAME = `
    SELECT
        id,
        name,
        description,
        slug,
        permissions,
        created_at,
        updated_at
    FROM roles
    WHERE name = $1;
`

var USER_ROLE_INSERT = `
    INSERT INTO user_roles (
        user_id,
        role_id
    )
    VALUES ($1, $2)
    ON CONFLICT (user_id, role_id) DO NOTHING;
`

var USER_ROLE_EXISTS = `
    SELECT EXISTS (
        SELECT 1
        FROM user_roles
        WHERE user_id = $1
          AND role_id = $2
    );
`

// ROLE_UPDATE_PERMISSIONS replaces the permission-ID array on a role.
// Role→permission links live in roles.permissions, not a join table.
var ROLE_UPDATE_PERMISSIONS = `
    UPDATE roles
    SET permissions = $2,
        updated_at = NOW()
    WHERE id = $1;
`

var PERMISSION_INSERT = `
	INSERT INTO permissions (
		id,
		name,
		code,
		"group",
        actions,
		slug,
		updated_by
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING
		id,
		name,
		code,
		"group",
        actions,
		slug,
		updated_by,
		created_at,
		updated_at,
		version;
`

var PERMISSION_FIND_BY_ID = `
	SELECT
		id,
		name,
		code,
		"group",
		actions,
		slug,
		updated_by,
		created_at,
		updated_at,
		version
	FROM permissions
	WHERE id = $1;
`

var PERMISSION_FIND_BY_NAME = `
	SELECT
		id,
		name,
		code,
		"group",
		actions,
		slug,
		updated_by,
		created_at,
		updated_at,
		version
	FROM permissions
	WHERE name = $1;
`

var PERMISSION_FIND_BY_GROUP = `
	SELECT
		id,
		name,
		code,
		"group",
		actions,
		slug,
		updated_by,
		created_at,
		updated_at,
		version
	FROM permissions
	WHERE "group" = $1
	ORDER BY name
	LIMIT 1;
`
