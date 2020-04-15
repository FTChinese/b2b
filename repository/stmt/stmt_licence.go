package stmt

const licenceCols = `l.id AS licence_id,
l.team_id AS team_id,
l.assignee_id AS assignee_id,
l.expire_date AS expire_date,
l.current_status AS current_status,
l.created_utc AS created_utc,
l.updated_utc AS updated_utc,
l.current_plan AS current_plan
l.last_invitation AS last_invitation`

const selectExpandedLicence = `
SELECT ` + licenceCols + `,
` + readerAccountCols + `
FROM b2b.licence AS l
LEFT JOIN cmstmp01.userinfo AS u
	ON l.assignee_id = u.user_id`

// Select a single licence belonging to a team.
const ExpandedLicence = selectExpandedLicence + `
WHERE l.id = ? AND l.team_id = ?
LIMIT 1`

// Select a list of licence for a team.
const ListExpandedLicences = selectExpandedLicence + `
WHERE l.team_id = ?
ORDER BY l.created_utc DESC
LIMIT ? OFFSET ?`

// CountLicence is used to support pagination.
const CountLicence = `
SELECT COUNT(*) AS total_licence
FROM b2b.licence
WHERE team_id = ?`

// LockLicence locks a row of licence
// when granting it to user.
const LockLicence = `
SELECT ` + licenceCols + `
FROM b2b.licence AS l
WHERE l.id = ?
LIMIT 1
FOR UPDATE`
