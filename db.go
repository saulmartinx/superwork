package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func selectStats() (*stats, error) {
	var result stats
	err := db.QueryRow(`
		SELECT
			(SELECT COUNT(1) FROM users WHERE deleted_at IS NULL) AS users,
			(SELECT COUNT(1) FROM companies WHERE deleted_at IS NULL) AS companies,
			(SELECT COUNT(1) FROM workflows WHERE deleted_at IS NULL) AS workflows,
			(SELECT COUNT(1) FROM organizations WHERE deleted_at IS NULL) AS organizations,
			(SELECT COUNT(1) FROM persons WHERE deleted_at IS NULL) AS people,
			(SELECT COUNT(1) FROM tasks WHERE deleted_at IS NULL) AS tasks,
			(SELECT COUNT(1) FROM activities WHERE deleted_at IS NULL) AS activities,
			(SELECT COUNT(1) FROM time_entries WHERE deleted_at IS NULL) AS time_entries
	`).Scan(
		&result.Users,
		&result.Companies,
		&result.Workflows,
		&result.Organizations,
		&result.People,
		&result.Tasks,
		&result.Activities,
		&result.TimeEntries,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func selectNoteByID(ID string) (*Note, error) {
	var model Note

	var personID sql.NullString
	var orgID sql.NullString
	var lastUpdateUserID sql.NullString
	var taskID sql.NullString

	err := db.QueryRow(`
		SELECT
		   	notes.id,
		   	notes.name,
			notes.company_id,
			notes.user_id,
			notes.task_id,
			notes.person_id,
			notes.org_id,
			notes.pinned_to_task_flag,
			notes.pinned_to_person_flag,
			notes.pinned_to_organization_flag,
			notes.last_update_user_id,
		    notes.created_at,
		    notes.updated_at,
		    notes.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name
		FROM
			notes
		LEFT OUTER JOIN
			users ON users.id = notes.user_id
		WHERE
			notes.id = $1
		LIMIT 1
	`,
		ID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.CompanyID,
		&model.UserID,
		&taskID,
		&personID,
		&orgID,
		&model.PinnedToTaskFlag,
		&model.PinnedToPersonFlag,
		&model.PinnedToOrganizationFlag,
		&lastUpdateUserID,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&model.UserName,
	)

	model.PersonID = personID.String
	model.OrgID = orgID.String
	model.LastUpdateUserID = lastUpdateUserID.String
	model.TaskID = taskID.String

	return &model, err
}

func selectWorkflowByID(workflowID string) (*Workflow, error) {
	var model Workflow

	var urlTitle sql.NullString
	var orderNr sql.NullInt64

	err := db.QueryRow(`
		SELECT
		   	id,
			company_id,
		   	name,
			order_nr,
			url_title,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			workflows
		WHERE
			id = $1
		LIMIT 1
	`,
		workflowID,
	).Scan(
		&model.ID,
		&model.CompanyID,
		&model.Name,
		&orderNr,
		&urlTitle,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)

	model.URLTitle = urlTitle.String
	model.OrderNr = int(orderNr.Int64)

	return &model, err
}

func selectUserByID(userID string) (*User, error) {
	var user User

	var activeCompanyID sql.NullString
	var activeWorkflowID sql.NullString
	var name sql.NullString
	var activeCompanyName sql.NullString
	var activeWorkflowName sql.NullString

	err := db.QueryRow(`
    	SELECT
    		users.id,
			users.name,
			users.email,
			users.password_hash,
			users.created_at,
			users.updated_at,
			users.deleted_at,
			users.phone,
			users.year_of_birth,
			users.picture,
			users.active_company_id,
			users.active_workflow_id,
			companies.name,
			workflows.name
    	FROM
    		users
    	LEFT OUTER JOIN
    		companies ON companies.id = users.active_company_id
    	LEFT OUTER JOIN
    		workflows ON workflows.id = users.active_workflow_id
    	WHERE
    		users.id = $1
    	LIMIT 1
    `,
		userID,
	).Scan(
		&user.ID,
		&name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Phone,
		&user.YearOfBirth,
		&user.Picture,
		&activeCompanyID,
		&activeWorkflowID,
		&activeCompanyName,
		&activeWorkflowName,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, errors.New("No user with that ID")
	case err != nil:
		return nil, err
	}

	user.ActiveCompanyID = activeCompanyID.String
	user.ActiveWorkflowID = activeWorkflowID.String
	user.Name = name.String
	if user.PasswordHash != "" {
		user.HasPassword = true
	}
	user.ActiveWorkflowName = activeWorkflowName.String
	user.ActiveCompanyName = activeCompanyName.String

	return &user, nil
}

func selectCompanyUserByUserAndCompany(userID, companyID string) (*CompanyUser, error) {
	var model CompanyUser

	err := db.QueryRow(`
		SELECT
			company_users.id,
			company_users.user_id,
			company_users.is_admin,
			company_users.company_id,
		    company_users.created_at,
		    company_users.updated_at,
		    company_users.deleted_at,
		    users.name,
		    users.email
		FROM
			company_users
		LEFT OUTER JOIN
			users ON users.id = company_users.user_id
		WHERE
			company_users.deleted_at IS NULL
		AND
			users.deleted_at IS NULL
		AND
			company_users.user_id = $1
		AND
			company_users.company_id = $2
		LIMIT 1
    `,
		userID,
		companyID,
	).Scan(
		&model.ID,
		&model.UserID,
		&model.IsAdmin,
		&model.CompanyID,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&model.Name,
		&model.Email,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	return &model, nil
}

func selectUserByEmail(email string) (*User, error) {
	var user User

	var activeCompanyID sql.NullString
	var activeWorkflowID sql.NullString
	var name sql.NullString
	var activeCompanyName sql.NullString
	var activeWorkflowName sql.NullString

	err := db.QueryRow(`
    	SELECT
    		users.id,
			users.name,
			users.email,
			users.password_hash,
			users.created_at,
			users.updated_at,
			users.deleted_at,
			users.phone,
			users.year_of_birth,
			users.picture,
			users.active_company_id,
			users.active_workflow_id,
			companies.name,
			workflows.name
    	FROM
    		users
    	LEFT OUTER JOIN
    		companies ON companies.id = users.active_company_id
    	LEFT OUTER JOIN
    		workflows ON workflows.id = users.active_workflow_id
    	WHERE
    		users.deleted_at is null
    	and
    		users.email = $1
    	LIMIT 1
    `,
		email,
	).Scan(
		&user.ID,
		&name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Phone,
		&user.YearOfBirth,
		&user.Picture,
		&activeCompanyID,
		&activeWorkflowID,
		&activeCompanyName,
		&activeWorkflowName,
	)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	user.ActiveCompanyID = activeCompanyID.String
	user.ActiveWorkflowID = activeWorkflowID.String
	user.Name = name.String
	if user.PasswordHash != "" {
		user.HasPassword = true
	}
	user.ActiveCompanyName = activeCompanyName.String
	user.ActiveWorkflowName = activeWorkflowName.String

	return &user, nil
}

func insertUser(model *User) error {
	if model.Name == "" {
		model.Name = model.Email
	}
	go sendEmail(
		config.AdminEmail,
		fmt.Sprintf("User %s (%s) signed up!", model.Name, model.Email),
		"")

	row := db.QueryRow(`
		INSERT INTO users(
			name,
			email,
			password_hash,
			picture,
			active_company_id,
			active_workflow_id,
			created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		maybeNull(model.Name),
		model.Email,
		model.PasswordHash,
		model.Picture,
		maybeNull(model.ActiveCompanyID),
		maybeNull(model.ActiveWorkflowID),
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateUser(model User) error {
	_, err := db.Exec(`
		UPDATE
			users
		SET
			phone = $1,
			year_of_birth = $2,
			updated_at = current_timestamp,
			name = $3,
			picture = $4,
			active_company_id = $5,
			active_workflow_id = $6
		WHERE
			id = $7
	`,
		model.Phone,
		model.YearOfBirth,
		model.Name,
		model.Picture,
		model.ActiveCompanyID,
		maybeNull(model.ActiveWorkflowID),
		model.ID,
	)
	if err != nil {
		return err
	}

	if model.PasswordHash != "" {
		_, err := db.Exec(`
			UPDATE
				users
			SET
				password_hash = $1
			WHERE
				id = $2
		`,
			model.PasswordHash,
			model.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteUser(ID string) error {
	_, err := db.Exec(`
		UPDATE
			users
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		ID,
	)
	return err
}

func insertCategory(model *Category) error {
	row := db.QueryRow(`
		INSERT INTO categories(
			name,
			created_at
		)
		VALUES(
			$1,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateCategory(model Category) error {
	_, err := db.Exec(`
		UPDATE
			categories
		SET
			name = $1,
			updated_at = current_timestamp
		WHERE
			id = $2
	`,
		model.Name,
		model.ID,
	)
	return err
}

func deleteCategory(model Category) error {
	_, err := db.Exec(`
		UPDATE
			categories
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertActivation(model *Activation) error {
	row := db.QueryRow(`
		INSERT INTO activations(
			user_id,
			created_at
		)
		VALUES(
			$1,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.UserID,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func selectActivationByID(activationID string) (*Activation, error) {
	var model Activation

	err := db.QueryRow(`
    	SELECT
    		id,
			user_id,
			created_at,
			updated_at,
			deleted_at
    	FROM
    		activations
    	WHERE
    		id = $1
    	LIMIT 1
    `,
		activationID,
	).Scan(
		&model.ID,
		&model.UserID,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)

	return &model, err
}

func deleteActivation(ID string) error {
	_, err := db.Exec(`
		UPDATE
			activations
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		ID,
	)
	return err
}

func insertCompany(model *Company) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	row := db.QueryRow(`
		INSERT INTO companies(
			name,
			created_at
		)
		VALUES(
			$1,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateCompany(model Company) error {
	_, err := db.Exec(`
		UPDATE
			companies
		SET
			name = $1,
			updated_at = current_timestamp
		WHERE
			id = $2
	`,
		model.Name,
		model.ID,
	)
	return err
}

func deleteCompany(ID string) error {
	_, err := db.Exec(`
		UPDATE
			companies
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		ID,
	)
	return err
}

func insertCompanyUser(model *CompanyUser) error {
	row := db.QueryRow(`
		INSERT INTO company_users(
			user_id,
			is_admin,
			company_id,
			created_at
		)
		VALUES(
			$1,
			true,
			$2,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.UserID,
		model.CompanyID,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateCompanyUser(model CompanyUser) error {
	_, err := db.Exec(`
		UPDATE
			company_users
		SET
			is_admin = $1,
			updated_at = current_timestamp
		WHERE
			id = $2
	`,
		model.IsAdmin,
		model.ID,
	)
	return err
}

func deleteCompanyUser(companyUserID string) error {
	_, err := db.Exec(`
		UPDATE
			company_users
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		companyUserID,
	)
	return err
}

func insertWorkflow(model *Workflow) error {
	if model.CompanyID == "" {
		return errors.New("Please select a company")
	}
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	row := db.QueryRow(`
		INSERT INTO workflows(
			company_id,
			name,
			order_nr,
			created_at
		)
		VALUES(
			$1,
			$2,
			0,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateWorkflow(model Workflow) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	_, err := db.Exec(`
		UPDATE
			workflows
		SET
			name = $1,
			order_nr = $2,
			updated_at = current_timestamp
		WHERE
			id = $3
	`,
		model.Name,
		model.OrderNr,
		model.ID,
	)
	return err
}

func workflowIsInUse(workflowID string) (bool, error) {
	var result bool
	err := db.QueryRow(`
		select
			count(1) > 0
		from
			tasks
		where
			deleted_at is null
		and
			workflow_id = $1
	`,
		workflowID,
	).Scan(
		&result,
	)
	return result, err
}

func deleteWorkflow(workflowID string) error {
	used, err := workflowIsInUse(workflowID)
	if err != nil {
		return err
	}
	if used {
		return errors.New("The workflow has tasks. Please remove the tasks first")
	}

	_, err = db.Exec(`
		UPDATE
			workflows
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		workflowID,
	)
	return err
}

func insertPerson(model *Person) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	row := db.QueryRow(`
		INSERT INTO persons(
			company_id,
			owner_id,
			org_id,
			first_name,
			name,
			created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.OwnerID,
		maybeNull(model.OrgID),
		model.FirstName,
		model.Name,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updatePerson(model Person) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	_, err := db.Exec(`
		UPDATE
			persons
		SET
			owner_id = $1,
			org_id = $2,
			first_name = $3,
			name = $4,
			updated_at = current_timestamp
		WHERE
			id = $5
	`,
		model.OwnerID,
		maybeNull(model.OrgID),
		model.FirstName,
		model.Name,
		model.ID,
	)
	return err
}

func deletePerson(personID string) error {
	_, err := db.Exec(`
		UPDATE
			persons
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		personID,
	)
	return err
}

func insertOrganization(model *Organization) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	row := db.QueryRow(`
		INSERT INTO organizations(
			company_id,
			owner_id,
			name,
			country_code,
			address,
			address_subpremise,
			address_street_number,
			address_route,
			address_sublocality,
			address_locality,
			address_admin_area_level_1,
			address_admin_area_level_2,
			address_country,
			address_postal_code,
			category_id,
			picture_id,
			first_char,
			visible_to,
			created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.OwnerID,
		model.Name,
		model.CountryCode,
		model.Address,
		model.AddressSubpremise,
		model.AddressStreetNumber,
		model.AddressRoute,
		model.AddressSublocality,
		model.AddressLocality,
		model.AddressAdminAreaLevel1,
		model.AddressAdminAreaLevel2,
		model.AddressCountry,
		model.AddressPostalCode,
		maybeNull(model.CategoryID),
		maybeNull(model.PictureID),
		model.FirstChar,
		model.VisibleTo,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateOrganization(model Organization) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	_, err := db.Exec(`
		UPDATE
			organizations
		SET
			owner_id = $1,
			name = $2,
			country_code = $3,
			address = $4,
			address_subpremise = $5,
			address_street_number = $6,
			address_route = $7,
			address_sublocality = $8,
			address_locality = $9,
			address_admin_area_level_1 = $10,
			address_admin_area_level_2 = $11,
			address_country = $12,
			address_postal_code = $13,
			category_id = $14,
			picture_id = $15,
			first_char = $16,
			visible_to = $17,
			updated_at = current_timestamp
		WHERE
			id = $18
	`,
		model.OwnerID,
		model.Name,
		model.CountryCode,
		model.Address,
		model.AddressSubpremise,
		model.AddressStreetNumber,
		model.AddressRoute,
		model.AddressSublocality,
		model.AddressLocality,
		model.AddressAdminAreaLevel1,
		model.AddressAdminAreaLevel2,
		model.AddressCountry,
		model.AddressPostalCode,
		maybeNull(model.CategoryID),
		maybeNull(model.PictureID),
		model.FirstChar,
		model.VisibleTo,
		model.ID,
	)
	return err
}

func deleteOrganization(organizationID string) error {
	_, err := db.Exec(`
		UPDATE
			organizations
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		organizationID,
	)
	return err
}

func insertActivity(model *Activity) error {
	if model.Name == "" {
		return errors.New("Please enter a subject")
	}

	row := db.QueryRow(`
		INSERT INTO activities(
			name,
			company_id,
			user_id,
			done,
			type_id,
			reference_type,
			reference_id,
			due_date,
			duration,
			marked_as_done_time,
			task_id,
			org_id,
			person_id,
			assigned_to_user_id,
			created_by_user_id,
			created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
		model.CompanyID,
		model.UserID,
		model.Done,
		maybeNull(model.TypeID),
		model.ReferenceType,
		model.ReferenceID,
		model.DueDate,
		model.Duration,
		model.DoneAt,
		maybeNull(model.TaskID),
		maybeNull(model.OrgID),
		maybeNull(model.PersonID),
		maybeNull(model.AssignedToUserID),
		maybeNull(model.CreatedByUserID),
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateActivity(model Activity) error {
	if model.Name == "" {
		return errors.New("Please enter a subject")
	}

	_, err := db.Exec(`
		UPDATE
			activities
		SET
			name = $1,
			user_id = $2,
			done = $3,
			type_id = $4,
			reference_type = $5,
			reference_id = $6,
			due_date = $7,
			duration = $8,
			marked_as_done_time = $9,
			task_id = $10,
			org_id = $11,
			person_id = $12,
			assigned_to_user_id = $13,
			created_by_user_id = $14,
			updated_at = current_timestamp
		WHERE
			id = $15
	`,
		model.Name,
		model.UserID,
		model.Done,
		maybeNull(model.TypeID),
		model.ReferenceType,
		model.ReferenceID,
		model.DueDate,
		model.Duration,
		model.DoneAt,
		maybeNull(model.TaskID),
		maybeNull(model.OrgID),
		maybeNull(model.PersonID),
		maybeNull(model.AssignedToUserID),
		maybeNull(model.CreatedByUserID),
		model.ID,
	)
	return err
}

func deleteActivity(activityID string) error {
	_, err := db.Exec(`
		UPDATE
			activities
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		activityID,
	)
	return err
}

func insertNote(model *Note) error {
	row := db.QueryRow(`
		INSERT INTO notes(
		  	name,
			company_id,
			user_id,
			task_id,
			person_id,
			org_id,
			pinned_to_task_flag,
			pinned_to_person_flag,
			pinned_to_organization_flag,
			last_update_user_id,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
		model.CompanyID,
		model.UserID,
		maybeNull(model.TaskID),
		maybeNull(model.PersonID),
		maybeNull(model.OrgID),
		model.PinnedToTaskFlag,
		model.PinnedToPersonFlag,
		model.PinnedToOrganizationFlag,
		maybeNull(model.LastUpdateUserID),
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateNote(model Note) error {
	_, err := db.Exec(`
		UPDATE
			notes
		SET
		  	name = $1,
			user_id = $2,
			task_id = $3,
			person_id = $4,
			org_id = $5,
			pinned_to_task_flag = $6,
			pinned_to_person_flag = $7,
			pinned_to_organization_flag = $8,
			last_update_user_id = $9,
			updated_at = current_timestamp
		WHERE
			id = $10
	`,
		model.Name,
		model.UserID,
		maybeNull(model.TaskID),
		maybeNull(model.PersonID),
		maybeNull(model.OrgID),
		model.PinnedToTaskFlag,
		model.PinnedToPersonFlag,
		model.PinnedToOrganizationFlag,
		maybeNull(model.LastUpdateUserID),
		model.ID,
	)
	return err
}

func deleteNote(noteID string) error {
	_, err := db.Exec(`
		UPDATE
			notes
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		noteID,
	)
	return err
}

func insertContact(model *Contact) error {
	if model.Name == "" {
		return errors.New("Please enter a phone number or an e-mail")
	}
	if model.Type == "" {
		return errors.New("Please select contact type")
	}

	row := db.QueryRow(`
		INSERT INTO contacts(
			name,
			person_id,
			type,
			primary_contact,
			created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
		model.PersonID,
		model.Type,
		model.Primary,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateContact(model Contact) error {
	if model.Name == "" {
		return errors.New("Please enter a phone number or an e-mail")
	}
	if model.Type == "" {
		return errors.New("Please select contact type")
	}

	_, err := db.Exec(`
		UPDATE
			contacts
		SET
			name = $1,
			type = $2,
			primary_contact = $3,
			updated_at = current_timestamp
		WHERE
			id = $4
	`,
		model.Name,
		model.Type,
		model.Primary,
		model.ID,
	)
	return err
}

func deleteContact(ID string) error {
	_, err := db.Exec(`
		UPDATE
			contacts
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		ID,
	)
	return err
}

func insertPersonField(model *PersonField) error {
	row := db.QueryRow(`
		INSERT INTO person_fields(
		   	company_id,
		   	name,
		   	key,
		   	order_nr,
		   	picklist_data,
			field_type,
			edit_flag,
			index_visible_flag,
			details_visible_flag,
			add_visible_flag,
			important_flag,
			bulk_edit_allowed,
			use_field,
			link,
			mandatory_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.UseField,
		model.Link,
		model.MandatoryFlag,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updatePersonField(model PersonField) error {
	_, err := db.Exec(`
		UPDATE
			person_fields
		SET
		   	name = $1,
		   	key = $2,
		   	order_nr = $3,
		   	picklist_data = $4,
			field_type = $5,
			edit_flag = $6,
			index_visible_flag = $7,
			details_visible_flag = $8,
			add_visible_flag = $9,
			important_flag = $10,
			bulk_edit_allowed = $11,
			use_field = $12,
			link = $13,
			mandatory_flag = $14,
			updated_at = current_timestamp
		WHERE
			id = $15
	`,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.UseField,
		model.Link,
		model.MandatoryFlag,
		model.ID,
	)
	return err
}

func deletePersonField(model PersonField) error {
	_, err := db.Exec(`
		UPDATE
			person_fields
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertFile(model *File) error {
	row := db.QueryRow(`
		INSERT INTO files(
		   	name,
			user_id,
			task_id,
			person_id,
			org_id,
			activity_id,
			note_id,
			file_type,
			file_size,
			inline_flag,
			remote_location,
			remote_id,
			cid,
			s3_bucket,
			url,
			description,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
		model.UserID,
		maybeNull(model.TaskID),
		maybeNull(model.PersonID),
		maybeNull(model.OrgID),
		maybeNull(model.ActivityID),
		maybeNull(model.NoteID),
		model.FileType,
		model.FileSize,
		model.InlineFlag,
		model.RemoteLocation,
		model.RemoteID,
		model.CID,
		model.S3Bucket,
		model.URL,
		model.Description,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateFile(model File) error {
	_, err := db.Exec(`
		UPDATE
			files
		SET
		   	name = $1,
			user_id = $2,
			task_id = $3,
			person_id = $4,
			org_id = $5,
			activity_id = $6,
			note_id = $7,
			file_type = $8,
			file_size = $9,
			inline_flag = $10,
			remote_location = $11,
			remote_id = $12,
			cid = $13,
			s3_bucket = $14,
			url = $15,
			description = $16,
			updated_at = current_timestamp
		WHERE
			id = $17
	`,
		model.Name,
		model.UserID,
		maybeNull(model.TaskID),
		maybeNull(model.PersonID),
		maybeNull(model.OrgID),
		maybeNull(model.ActivityID),
		maybeNull(model.NoteID),
		model.FileType,
		model.FileSize,
		model.InlineFlag,
		model.RemoteLocation,
		model.RemoteID,
		model.CID,
		model.S3Bucket,
		model.URL,
		model.Description,
		model.ID,
	)
	return err
}

func deleteFile(model File) error {
	_, err := db.Exec(`
		UPDATE
			files
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertStage(model *Stage) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	row := db.QueryRow(`
		INSERT INTO stages(
		  	name,
			order_nr,
			task_probability,
			workflow_id,
			rotten_flag,
			rotten_days,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
		model.OrderNr,
		model.TaskProbability,
		model.WorkflowID,
		model.RottenFlag,
		model.RottenDays,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateStage(model Stage) error {
	if model.Name == "" {
		return errors.New("Please enter a name")
	}

	_, err := db.Exec(`
		UPDATE
			stages
		SET
		  	name = $1,
			order_nr = $2,
			task_probability = $3,
			workflow_id = $4,
			rotten_flag = $5,
			rotten_days = $6,
			updated_at = current_timestamp
		WHERE
			id = $7
	`,
		model.Name,
		model.OrderNr,
		model.TaskProbability,
		model.WorkflowID,
		model.RottenFlag,
		model.RottenDays,
		model.ID,
	)
	return err
}

func stageIsInUse(stageID string) (bool, error) {
	var result bool
	err := db.QueryRow(`
		select
			count(1) > 0
		from
			tasks
		where
			deleted_at is null
		and
			stage_id = $1
	`,
		stageID,
	).Scan(
		&result,
	)
	return result, err
}

func deleteStage(stageID string) error {
	used, err := stageIsInUse(stageID)
	if err != nil {
		return err
	}
	if used {
		return errors.New("The stage has tasks. Please remove the tasks first")
	}

	_, err = db.Exec(`
		UPDATE
			stages
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		stageID,
	)
	return err
}

func insertTask(model *Task) error {
	if model.CompanyID == "" {
		return errors.New("Please assign a company")
	}
	if model.StageID == "" {
		return errors.New("Please assign a workflow stage")
	}
	if model.Name == "" {
		return errors.New("Please enter a title")
	}

	row := db.QueryRow(`
		INSERT INTO tasks(
		   	name,
			creator_user_id,
			user_id,
			person_id,
			org_id,
			stage_id,
			value,
			currency,
			stage_change_time,
			status,
			lost_reason,
			visible_to,
			close_time,
			workflow_id,
			won_time,
			first_won_time,
			lost_time,
			expected_close_date,
			stage_order_nr,
			formatted_value,
			rotten_time,
			weighted_value,
			formatted_weighted_value,
			cc_email,
			org_hidden,
			person_hidden,
			company_id,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			$19,
			$20,
			$21,
			$22,
			$23,
			$24,
			$25,
			$26,
			$27,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.Name,
		model.CreatorUserID,
		model.UserID,
		maybeNull(model.PersonID),
		maybeNull(model.OrgID),
		model.StageID,
		model.Value,
		model.Currency,
		model.StageChangeTime,
		model.Status,
		model.LostReason,
		model.VisibleTo,
		model.CloseTime,
		model.WorkflowID,
		model.WonTime,
		model.FirstWonTime,
		model.LostTime,
		model.ExpectedCloseDate,
		model.StageOrderNr,
		model.FormattedValue,
		model.RottenTime,
		model.WeightedValue,
		model.FormattedWeightedValue,
		model.CCEmail,
		model.OrgHidden,
		model.PersonHidden,
		model.CompanyID,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateTask(model Task) error {
	if model.CompanyID == "" {
		return errors.New("Please assign a company")
	}
	if model.StageID == "" {
		return errors.New("Please assign a workflow stage")
	}
	if model.Name == "" {
		return errors.New("Please enter a title")
	}

	_, err := db.Exec(`
		UPDATE
			tasks
		SET
		   	name = $1,
			creator_user_id = $2,
			user_id = $3,
			person_id = $4,
			org_id = $5,
			stage_id = $6,
			value = $7,
			currency = $8,
			stage_change_time = $9,
			status = $10,
			lost_reason = $11,
			visible_to = $12,
			close_time = $13,
			workflow_id = $14,
			won_time = $15,
			first_won_time = $16,
			lost_time = $17,
			expected_close_date = $18,
			stage_order_nr = $19,
			formatted_value = $20,
			rotten_time = $21,
			weighted_value = $22,
			formatted_weighted_value = $23,
			cc_email = $24,
			org_hidden = $25,
			person_hidden = $26,
			updated_at = current_timestamp
		WHERE
			id = $27
	`,
		model.Name,
		model.CreatorUserID,
		model.UserID,
		maybeNull(model.PersonID),
		maybeNull(model.OrgID),
		model.StageID,
		model.Value,
		model.Currency,
		model.StageChangeTime,
		model.Status,
		model.LostReason,
		model.VisibleTo,
		model.CloseTime,
		model.WorkflowID,
		model.WonTime,
		model.FirstWonTime,
		model.LostTime,
		model.ExpectedCloseDate,
		model.StageOrderNr,
		model.FormattedValue,
		model.RottenTime,
		model.WeightedValue,
		model.FormattedWeightedValue,
		model.CCEmail,
		model.OrgHidden,
		model.PersonHidden,
		model.ID,
	)
	return err
}

func deleteTask(taskID string) error {
	_, err := db.Exec(`
		UPDATE
			tasks
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		taskID,
	)
	return err
}

func insertOrganizationField(model *OrganizationField) error {
	row := db.QueryRow(`
		INSERT INTO organization_fields(
			company_id,
			name,
			key,
			order_nr,
			picklist_data,
			field_type,
			edit_flag,
			index_visible_flag,
			details_visible_flag,
			add_visible_flag,
			important_flag,
			bulk_edit_allowed,
			use_field,
			link,
			mandatory_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.UseField,
		model.Link,
		model.MandatoryFlag,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateOrganizationField(model OrganizationField) error {
	_, err := db.Exec(`
		UPDATE
			organization_fields
		SET
			name = $1,
			key = $2,
			order_nr = $3,
			picklist_data = $4,
			field_type = $5,
			edit_flag = $6,
			index_visible_flag = $7,
			details_visible_flag = $8,
			add_visible_flag = $9,
			important_flag = $10,
			bulk_edit_allowed = $11,
			use_field = $12,
			link = $13,
			mandatory_flag = $14,
			updated_at = current_timestamp
		WHERE
			id = $15
	`,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.UseField,
		model.Link,
		model.MandatoryFlag,
		model.ID,
	)
	return err
}

func deleteOrganizationField(model OrganizationField) error {
	_, err := db.Exec(`
		UPDATE
			organization_fields
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertTaskField(model *TaskField) error {
	row := db.QueryRow(`
		INSERT INTO task_fields(
			company_id,
			name,
			key,
			order_nr,
			picklist_data,
			field_type,
			edit_flag,
			index_visible_flag,
			details_visible_flag,
			add_visible_flag,
			important_flag,
			bulk_edit_allowed,
			mandatory_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.MandatoryFlag,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateTaskField(model TaskField) error {
	_, err := db.Exec(`
		UPDATE
			task_fields
		SET
			name = $1,
			key = $2,
			order_nr = $3,
			picklist_data = $4,
			field_type = $5,
			edit_flag = $6,
			index_visible_flag = $7,
			details_visible_flag = $8,
			add_visible_flag = $9,
			important_flag = $10,
			bulk_edit_allowed = $11,
			mandatory_flag = $12,
			updated_at = current_timestamp
		WHERE
			id = $13
	`,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.MandatoryFlag,
		model.ID,
	)
	return err
}

func deleteTaskField(model TaskField) error {
	_, err := db.Exec(`
		UPDATE
			task_fields
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func selectTasksByWorkflow(workflowID string, activeOnly bool) ([]Task, error) {
	if workflowID == "" {
		return nil, nil
	}
	rows, err := db.Query(`
		SELECT
		   	tasks.id,
		   	tasks.name,
			tasks.creator_user_id,
			tasks.user_id,
			tasks.person_id,
			tasks.org_id,
			tasks.stage_id,
			tasks.value,
			tasks.currency,
			tasks.stage_change_time,
			tasks.status,
			tasks.lost_reason,
			tasks.visible_to,
			tasks.close_time,
			tasks.workflow_id,
			tasks.won_time,
			tasks.first_won_time,
			tasks.lost_time,
			tasks.expected_close_date,
			tasks.stage_order_nr,
			tasks.formatted_value,
			tasks.rotten_time,
			tasks.weighted_value,
			tasks.formatted_weighted_value,
			tasks.cc_email,
			tasks.org_hidden,
			tasks.person_hidden,
		    tasks.created_at,
		    tasks.updated_at,
		    tasks.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		   	persons.name as person_name,
		   	organizations.name as org_name,
		   	stages.name as stage_name,
		   	(
		   		select activities.due_date
		   		from activities
		   		where activities.task_id = tasks.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_date,
		   	(
		   		select activities.id
		   		from activities
		   		where activities.task_id = tasks.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_id,
		   	tasks.company_id
		FROM
			tasks
		LEFT OUTER JOIN
			users ON users.id = tasks.user_id
		LEFT OUTER JOIN
			persons ON persons.id = tasks.person_id
		LEFT OUTER JOIN
			organizations ON organizations.id = tasks.org_id
		LEFT OUTER JOIN
			stages ON stages.id = tasks.stage_id
		WHERE
			tasks.deleted_at IS NULL
		AND
			tasks.workflow_id = $1
		AND (
			($2 AND tasks.won_time IS NULL AND tasks.lost_time IS NULL) OR (NOT $2)
		)
	`,
		workflowID,
		activeOnly,
	)
	if err != nil {
		return nil, err
	}

	return scanTasks(rows)
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	defer rows.Close()
	var result []Task
	for rows.Next() {

		var model Task

		var personID sql.NullString
		var personName sql.NullString
		var orgID sql.NullString
		var orgName sql.NullString
		var ownerName sql.NullString
		var nextActivityID sql.NullString

		if err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.CreatorUserID,
			&model.UserID,
			&personID,
			&orgID,
			&model.StageID,
			&model.Value,
			&model.Currency,
			&model.StageChangeTime,
			&model.Status,
			&model.LostReason,
			&model.VisibleTo,
			&model.CloseTime,
			&model.WorkflowID,
			&model.WonTime,
			&model.FirstWonTime,
			&model.LostTime,
			&model.ExpectedCloseDate,
			&model.StageOrderNr,
			&model.FormattedValue,
			&model.RottenTime,
			&model.WeightedValue,
			&model.FormattedWeightedValue,
			&model.CCEmail,
			&model.OrgHidden,
			&model.PersonHidden,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
			&ownerName,
			&personName,
			&orgName,
			&model.StageName,
			&model.NextActivityDate,
			&nextActivityID,
			&model.CompanyID,
		); err != nil {
			return nil, err
		}

		model.PersonID = personID.String
		model.PersonName = personName.String
		model.OrgID = orgID.String
		model.OrgName = orgName.String
		model.OwnerName = ownerName.String
		model.NextActivityID = nextActivityID.String

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectTasksByPerson(personID string) ([]Task, error) {
	rows, err := db.Query(`
		SELECT
		   	tasks.id,
		   	tasks.name,
			tasks.creator_user_id,
			tasks.user_id,
			tasks.person_id,
			tasks.org_id,
			tasks.stage_id,
			tasks.value,
			tasks.currency,
			tasks.stage_change_time,
			tasks.status,
			tasks.lost_reason,
			tasks.visible_to,
			tasks.close_time,
			tasks.workflow_id,
			tasks.won_time,
			tasks.first_won_time,
			tasks.lost_time,
			tasks.expected_close_date,
			tasks.stage_order_nr,
			tasks.formatted_value,
			tasks.rotten_time,
			tasks.weighted_value,
			tasks.formatted_weighted_value,
			tasks.cc_email,
			tasks.org_hidden,
			tasks.person_hidden,
		    tasks.created_at,
		    tasks.updated_at,
		    tasks.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		   	persons.name as person_name,
		   	organizations.name as org_name,
		   	stages.name as stage_name,
		   	(
		   		select activities.due_date
		   		from activities
		   		where activities.task_id = tasks.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_date,
		   	(
		   		select activities.id
		   		from activities
		   		where activities.task_id = tasks.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_id,
		   	tasks.company_id
		FROM
			tasks
		LEFT OUTER JOIN
			users ON users.id = tasks.user_id
		LEFT OUTER JOIN
			persons ON persons.id = tasks.person_id
		LEFT OUTER JOIN
			organizations ON organizations.id = tasks.org_id
		LEFT OUTER JOIN
			stages ON stages.id = tasks.stage_id
		WHERE
			tasks.deleted_at IS NULL
		AND
			tasks.person_id = $1
	`,
		personID,
	)
	if err != nil {
		return nil, err
	}

	return scanTasks(rows)
}

func selectFirstWorkflowByCompany(companyID string) (ID string, name string, err error) {
	err = db.QueryRow(`
		select
			id,
			name
		from
			workflows
		where
			deleted_at is null
		and
			company_id = $1
		limit 1
	`,
		companyID,
	).Scan(
		&ID,
		&name,
	)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return ID, name, err
}

func selectFirstCompanyByUser(userID string) (string, error) {
	var result string
	err := db.QueryRow(`
		select
			company_id
		from
			company_users
		where
			deleted_at is null
		and
			user_id = $1
		limit 1
	`,
		userID,
	).Scan(
		&result,
	)
	return result, err
}

func selectActivityTypeByID(ID string) (*ActivityType, error) {
	var model ActivityType
	err := db.QueryRow(`
		select
		   	id,
			company_id,
		   	name,
			key_string,
			order_nr,
			color,
			is_custom_flag,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			activity_types
		WHERE
			id = $1
	`,
		ID,
	).Scan(
		&model.ID,
		&model.CompanyID,
		&model.Name,
		&model.KeyString,
		&model.OrderNr,
		&model.Color,
		&model.IsCustomFlag,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)
	return &model, err
}

func selectActivityTypesByCompany(companyID string) ([]ActivityType, error) {
	rows, err := db.Query(`
		SELECT
		   	id,
			company_id,
		   	name,
			key_string,
			order_nr,
			color,
			is_custom_flag,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			activity_types
		WHERE
			deleted_at IS NULL
		AND
			company_id = $1
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ActivityType
	for rows.Next() {

		var model ActivityType

		err = rows.Scan(
			&model.ID,
			&model.CompanyID,
			&model.Name,
			&model.KeyString,
			&model.OrderNr,
			&model.Color,
			&model.IsCustomFlag,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectWorkflowsByCompany(companyID string) ([]Workflow, error) {
	rows, err := db.Query(`
		SELECT
		   	id,
			company_id,
		   	name,
			order_nr,
			url_title,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			workflows
		WHERE
			deleted_at IS NULL
		AND
			company_id = $1
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Workflow
	for rows.Next() {

		var model Workflow

		var urlTitle sql.NullString
		var orderNr sql.NullInt64

		err = rows.Scan(
			&model.ID,
			&model.CompanyID,
			&model.Name,
			&orderNr,
			&urlTitle,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		model.URLTitle = urlTitle.String
		model.OrderNr = int(orderNr.Int64)

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectActivitiesByCompany(companyID string) ([]Activity, error) {
	rows, err := db.Query(`
		SELECT
		   	activities.id,
		   	activities.name,
			activities.company_id,
			activities.user_id,
			activities.done,
			activities.reference_type,
			activities.reference_id,
			activities.due_date,
			activities.duration,
			activities.marked_as_done_time,
			activities.task_id,
			activities.org_id,
			activities.person_id,
			activities.assigned_to_user_id,
			activities.created_by_user_id,
		    activities.created_at,
		    activities.updated_at,
		    activities.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name,
		    persons.name as person_name,
		    tasks.name as task_name,
		    organizations.name as org_name,
		    activities.type_id,
		    activity_types.name as type,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as person_email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as person_phone,
		   	assigned_users.name as assigned_to_user_name
		FROM
			activities
		LEFT OUTER JOIN
			users ON users.id = activities.user_id
		LEFT OUTER JOIN
			users AS assigned_users ON assigned_users.id = activities.assigned_to_user_id
		LEFT OUTER JOIN
			persons ON persons.id = activities.person_id
		LEFT OUTER JOIN
			tasks ON tasks.id = activities.task_id
		LEFT OUTER JOIN
			organizations ON organizations.id = activities.org_id
		LEFT OUTER JOIN
			activity_types ON activity_types.id = activities.type_id
		WHERE
			activities.deleted_at IS NULL
		AND
			activities.company_id = $1
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}

	return scanActivities(rows)
}

func selectCompanyUsersByUser(userID string) ([]CompanyUser, error) {
	rows, err := db.Query(`
		SELECT
			company_users.id,
			company_users.user_id,
			company_users.is_admin,
			company_users.company_id,
		    company_users.created_at,
		    company_users.updated_at,
		    company_users.deleted_at,
		    users.email,
		    users.name
		FROM
			company_users
		LEFT OUTER JOIN
			users ON users.id = company_users.user_id
		WHERE
			company_users.deleted_at IS NULL
		AND
			company_users.user_id = $1
	`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return scanCompanyUsers(rows)
}

func selectContactByID(ID string) (*Contact, error) {
	var model Contact

	err := db.QueryRow(`
		SELECT
		   	id,
		   	name,
			person_id,
			primary_contact,
			type,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			contacts
		WHERE
			id = $1
		LIMIT 1
	`,
		ID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.PersonID,
		&model.Primary,
		&model.Type,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, errors.New("No contact with that ID")
	case err != nil:
		return nil, err
	}

	return &model, nil
}

func selectContactsByPerson(personID string) ([]Contact, error) {
	rows, err := db.Query(`
		SELECT
		   	id,
		   	name,
			person_id,
			primary_contact,
			type,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			contacts
		WHERE
			person_id = $1
		AND
			deleted_at is null
		ORDER BY
			type, name
	`,
		personID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Contact
	for rows.Next() {
		var model Contact

		err = rows.Scan(
			&model.ID,
			&model.Name,
			&model.PersonID,
			&model.Primary,
			&model.Type,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectPersonsByCompany(companyID string) ([]Person, error) {
	rows, err := db.Query(`
		SELECT
		   	persons.id,
		   	persons.name,
			persons.company_id,
			persons.owner_id,
			persons.org_id,
			persons.first_name,
		    persons.created_at,
		    persons.updated_at,
		    persons.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.person_id = persons.id
		    	and deleted_at is null
		    	and (won_time is null and lost_time is null)
		    ) as open_tasks_count,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.person_id = persons.id
		    	and deleted_at is null
		    	and (won_time is not null or lost_time is not null)
		    ) as closed_tasks_count,
		    organizations.name as org_name,
		   	(
		   		select activities.due_date
		   		from activities
		   		where activities.person_id = persons.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_date,
		   	(
		   		select activities.id
		   		from activities
		   		where activities.person_id = persons.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_id,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as phone
		FROM
			persons
		LEFT OUTER JOIN
			users ON users.id = persons.owner_id
		LEFT OUTER JOIN
			organizations ON organizations.id = persons.org_id
		WHERE
			persons.deleted_at IS NULL
		AND
			persons.company_id = $1
		ORDER BY
			persons.name
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}

	return scanPersons(rows)
}

func scanPersons(rows *sql.Rows) ([]Person, error) {
	defer rows.Close()
	var result []Person
	for rows.Next() {
		var model Person

		var orgID sql.NullString
		var ownerName sql.NullString
		var orgName sql.NullString
		var nextActivityID sql.NullString
		var email sql.NullString
		var phone sql.NullString

		if err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.CompanyID,
			&model.OwnerID,
			&orgID,
			&model.FirstName,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
			&ownerName,
			&model.OpenTasksCount,
			&model.ClosedTasksCount,
			&orgName,
			&model.NextActivityDate,
			&nextActivityID,
			&email,
			&phone,
		); err != nil {
			return nil, err
		}

		model.OrgID = orgID.String
		model.OwnerName = ownerName.String
		model.OrgName = orgName.String
		model.NextActivityID = nextActivityID.String
		model.Email = email.String
		model.Phone = phone.String

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectCurrenciesByCompany(companyID string) ([]Currency, error) {
	rows, err := db.Query(`
		SELECT
		   	id,
			company_id,
			name,
			code,
			decimal_points,
			symbol,
			is_custom_flag,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			currencies
		WHERE
			deleted_at IS NULL
		AND
			company_id = $1
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Currency
	for rows.Next() {
		var model Currency

		err = rows.Scan(
			&model.ID,
			&model.CompanyID,
			&model.Name,
			&model.Code,
			&model.DecimalPoints,
			&model.Symbol,
			&model.IsCustomFlag,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectOrganizationsByCompany(companyID string) ([]Organization, error) {
	rows, err := db.Query(`
		SELECT
		   	organizations.id,
		   	organizations.name,
			organizations.company_id,
			organizations.owner_id,
			organizations.country_code,
			organizations.address,
			organizations.address_subpremise,
			organizations.address_street_number,
			organizations.address_route,
			organizations.address_sublocality,
			organizations.address_locality,
			organizations.address_admin_area_level_1,
			organizations.address_admin_area_level_2,
			organizations.address_country,
			organizations.address_postal_code,
			organizations.category_id,
			organizations.first_char,
			organizations.visible_to,
		    organizations.created_at,
		    organizations.updated_at,
		    organizations.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.org_id = organizations.id
		    	and deleted_at is null
		    	and (won_time is null and lost_time is null)
		    ) as open_tasks_count,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.org_id = organizations.id
		    	and deleted_at is null
		    	and (won_time is not null or lost_time is not null)
		    ) as closed_tasks_count,
		    (
		    	select count(1)
		    	from persons
		    	where persons.org_id = organizations.id
		    	and persons.deleted_at is null
		    ) as people_count,
		    (
		    	select min(due_date)
		    	from activities
		    	where deleted_at is null
		    	and activities.org_id = organizations.id
		    ) as next_activity_date,
		    (
		    	select activities.id
		    	from activities
		    	where deleted_at is null
		    	and activities.org_id = organizations.id
		    	order by activities.due_date asc
		    	limit 1
		    ) as next_activity_id
		FROM
			organizations
		LEFT OUTER JOIN
			users ON users.id = organizations.owner_id
		WHERE
			organizations.deleted_at IS NULL
		AND
			organizations.company_id = $1
		ORDER BY
			organizations.name
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Organization
	for rows.Next() {
		model, err := scanOrganization(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *model)
	}
	return result, rows.Err()
}

func scanOrganization(rows *sql.Rows) (*Organization, error) {
	var model Organization

	var categoryID sql.NullString
	var ownerName sql.NullString
	var nextActivityID sql.NullString

	if err := rows.Scan(
		&model.ID,
		&model.Name,
		&model.CompanyID,
		&model.OwnerID,
		&model.CountryCode,
		&model.Address,
		&model.AddressSubpremise,
		&model.AddressStreetNumber,
		&model.AddressRoute,
		&model.AddressSublocality,
		&model.AddressLocality,
		&model.AddressAdminAreaLevel1,
		&model.AddressAdminAreaLevel2,
		&model.AddressCountry,
		&model.AddressPostalCode,
		&categoryID,
		&model.FirstChar,
		&model.VisibleTo,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&ownerName,
		&model.OpenTasksCount,
		&model.ClosedTasksCount,
		&model.PeopleCount,
		&model.NextActivityDate,
		&nextActivityID,
	); err != nil {
		return nil, err
	}

	model.CategoryID = categoryID.String
	model.OwnerName = ownerName.String
	model.NextActivityID = nextActivityID.String

	return &model, nil
}

func selectNotesByTask(TaskID string) ([]Note, error) {
	rows, err := db.Query(`
		SELECT
		   	notes.id,
		   	notes.name,
			notes.company_id,
			notes.user_id,
			notes.task_id,
			notes.person_id,
			notes.org_id,
			notes.pinned_to_task_flag,
			notes.pinned_to_person_flag,
			notes.pinned_to_organization_flag,
			notes.last_update_user_id,
		    notes.created_at,
		    notes.updated_at,
		    notes.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name
		FROM
			notes
		LEFT OUTER JOIN
			users ON users.id = notes.user_id
		WHERE
			notes.deleted_at IS NULL
		AND
			notes.task_id = $1
	`,
		TaskID,
	)
	if err != nil {
		return nil, err
	}

	return scanNotes(rows)
}

func scanNotes(rows *sql.Rows) ([]Note, error) {
	defer rows.Close()
	var result []Note
	for rows.Next() {
		var model Note

		var personID sql.NullString
		var orgID sql.NullString
		var lastUpdateUserID sql.NullString
		var taskID sql.NullString

		if err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.CompanyID,
			&model.UserID,
			&taskID,
			&personID,
			&orgID,
			&model.PinnedToTaskFlag,
			&model.PinnedToPersonFlag,
			&model.PinnedToOrganizationFlag,
			&lastUpdateUserID,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
			&model.UserName,
		); err != nil {
			return nil, err
		}

		model.PersonID = personID.String
		model.OrgID = orgID.String
		model.LastUpdateUserID = lastUpdateUserID.String
		model.TaskID = taskID.String

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectNotesByPerson(personID string) ([]Note, error) {
	rows, err := db.Query(`
		SELECT
		   	notes.id,
		   	notes.name,
			notes.company_id,
			notes.user_id,
			notes.task_id,
			notes.person_id,
			notes.org_id,
			notes.pinned_to_task_flag,
			notes.pinned_to_person_flag,
			notes.pinned_to_organization_flag,
			notes.last_update_user_id,
		    notes.created_at,
		    notes.updated_at,
		    notes.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name
		FROM
			notes
		LEFT OUTER JOIN
			users ON users.id = notes.user_id
		WHERE
			notes.deleted_at IS NULL
		AND
			notes.person_id = $1
	`,
		personID,
	)
	if err != nil {
		return nil, err
	}

	return scanNotes(rows)
}

func selectPersonsByOrganization(organizationID string) ([]Person, error) {
	rows, err := db.Query(`
		SELECT
		   	persons.id,
		   	persons.name,
			persons.company_id,
			persons.owner_id,
			persons.org_id,
			persons.first_name,
		    persons.created_at,
		    persons.updated_at,
		    persons.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.person_id = persons.id
		    	and deleted_at is null
		    	and (won_time is null and lost_time is null)
		    ) as open_tasks_count,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.person_id = persons.id
		    	and deleted_at is null
		    	and (won_time is not null or lost_time is not null)
		    ) as closed_tasks_count,
		    organizations.name as org_name,
		   	(
		   		select activities.due_date
		   		from activities
		   		where activities.person_id = persons.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_date,
		   	(
		   		select activities.id
		   		from activities
		   		where activities.person_id = persons.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_id,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as phone
		FROM
			persons
		LEFT OUTER JOIN
			users ON users.id = persons.owner_id
		LEFT OUTER JOIN
			organizations ON organizations.id = persons.org_id
		WHERE
			persons.deleted_at IS NULL
		AND
			persons.org_id = $1
		ORDER BY
			persons.name
	`,
		organizationID,
	)
	if err != nil {
		return nil, err
	}

	return scanPersons(rows)
}

func selectStageByID(stageID string) (*Stage, error) {
	var model Stage
	err := db.QueryRow(`
		select
		   	id,
		   	name,
			order_nr,
			task_probability,
			workflow_id,
			rotten_flag,
			rotten_days,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			stages
		WHERE
			id = $1
	`,
		stageID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.OrderNr,
		&model.TaskProbability,
		&model.WorkflowID,
		&model.RottenFlag,
		&model.RottenDays,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)
	return &model, err
}

func selectStagesByWorkflow(workflowID string) ([]Stage, error) {
	if workflowID == "" {
		return nil, nil
	}
	rows, err := db.Query(`
		SELECT
		   	id,
		   	name,
			order_nr,
			task_probability,
			workflow_id,
			rotten_flag,
			rotten_days,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			stages
		WHERE
			deleted_at IS NULL
		AND
			workflow_id = $1
		ORDER BY
			order_nr
	`,
		workflowID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Stage
	for rows.Next() {
		var model Stage
		err = rows.Scan(
			&model.ID,
			&model.Name,
			&model.OrderNr,
			&model.TaskProbability,
			&model.WorkflowID,
			&model.RottenFlag,
			&model.RottenDays,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, model)
	}
	return result, rows.Err()
}

func insertActivityField(model *ActivityField) error {
	row := db.QueryRow(`
		INSERT INTO activity_fields(
			company_id,
			name,
			key,
			order_nr,
			picklist_data,
			field_type,
			edit_flag,
			index_visible_flag,
			details_visible_flag,
			add_visible_flag,
			important_flag,
			bulk_edit_allowed,
			mandatory_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.MandatoryFlag,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateActivityField(model ActivityField) error {
	_, err := db.Exec(`
		UPDATE
			activity_fields
		SET
			name = $1,
			key = $2,
			order_nr = $3,
			picklist_data = $4,
			field_type = $5,
			edit_flag = $6,
			index_visible_flag = $7,
			details_visible_flag = $8,
			add_visible_flag = $9,
			important_flag = $10,
			bulk_edit_allowed = $11,
			mandatory_flag = $12,
			updated_at = current_timestamp
		WHERE
			id = $13
	`,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.MandatoryFlag,
		model.ID,
	)
	return err
}

func deleteActivityField(model ActivityField) error {
	_, err := db.Exec(`
		UPDATE
			activity_fields
		SET
			updated_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertActivityType(model *ActivityType) error {
	row := db.QueryRow(`
		INSERT INTO activity_types(
			company_id,
			name,
			key_string,
			order_nr,
			color,
			is_custom_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.KeyString,
		model.OrderNr,
		model.Color,
		model.IsCustomFlag,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateActivityType(model ActivityType) error {
	_, err := db.Exec(`
		UPDATE
			activity_types
		SET
			name = $1,
			key_string = $2,
			order_nr = $3,
			color = $4,
			is_custom_flag = $5,
			updated_at = current_timestamp
		WHERE
			id = $6
	`,
		model.Name,
		model.KeyString,
		model.OrderNr,
		model.Color,
		model.IsCustomFlag,
		model.ID,
	)
	return err
}

func deleteActivityType(model ActivityType) error {
	_, err := db.Exec(`
		UPDATE
			activity_types
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertCurrency(model *Currency) error {
	row := db.QueryRow(`
		INSERT INTO currencies(
			company_id,
			name,
			code,
			decimal_points,
			symbol,
			is_custom_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.Code,
		model.DecimalPoints,
		model.Symbol,
		model.IsCustomFlag,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateCurrency(model Currency) error {
	_, err := db.Exec(`
		UPDATE
			currencies
		SET
			name = $1,
			code = $2,
			decimal_points = $3,
			symbol = $4,
			is_custom_flag = $5,
			updated_at = current_timestamp
		WHERE
			id = $6
	`,
		model.Name,
		model.Code,
		model.DecimalPoints,
		model.Symbol,
		model.IsCustomFlag,
		model.ID,
	)
	return err
}

func deleteCurrency(model Currency) error {
	_, err := db.Exec(`
		UPDATE
			currencies
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertFilter(model *Filter) error {
	row := db.QueryRow(`
		INSERT INTO filters(
			company_id,
			name,
			type,
			temporary_flag,
			user_id,
			visible_to,
			custom_view_id,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.Type,
		model.TemporaryFlag,
		model.UserID,
		model.VisibleTo,
		model.CustomViewID,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateFilter(model Filter) error {
	_, err := db.Exec(`
		UPDATE
			filters
		SET
			name = $1,
			type = $2,
			temporary_flag = $3,
			user_id = $4,
			visible_to = $5,
			custom_view_id = $6,
			updated_at = current_timestamp
		WHERE
			id = $7
	`,
		model.Name,
		model.Type,
		model.TemporaryFlag,
		model.UserID,
		model.VisibleTo,
		model.CustomViewID,
		model.ID,
	)
	return err
}

func deleteFilter(model Filter) error {
	_, err := db.Exec(`
		UPDATE
			filters
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertGoal(model *Goal) error {
	row := db.QueryRow(`
		INSERT INTO goals(
			company_id,
			name,
			user_id,
			stage_id,
			active_goal_id,
			period,
			expected,
			goal_type,
			expected_sum,
			currency,
			expected_type,
			created_by_user_id,
			workflow_id,
			master_expected,
			delivered,
			delivered_sum,
			period_start,
			period_end,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.UserID,
		maybeNull(model.StageID),
		maybeNull(model.ActiveGoalID),
		model.Period,
		model.Expected,
		model.GoalType,
		model.ExpectedSum,
		model.Currency,
		model.ExpectedType,
		maybeNull(model.CreatedByUserID),
		maybeNull(model.WorkflowID),
		model.MasterExpected,
		model.Delivered,
		model.DeliveredSum,
		model.PeriodStart,
		model.PeriodEnd,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateGoal(model Goal) error {
	_, err := db.Exec(`
		UPDATE
			goals
		SET
			name = $1,
			user_id = $2,
			stage_id = $3,
			active_goal_id = $4,
			period = $5,
			expected = $6,
			goal_type = $7,
			expected_sum = $8,
			currency = $9,
			expected_type = $10,
			created_by_user_id = $11,
			workflow_id = $12,
			master_expected = $13,
			delivered = $14,
			delivered_sum = $15,
			period_start = $16,
			period_end = $17,
			updated_at = current_timestamp
		WHERE
			id = $18
	`,
		model.Name,
		model.UserID,
		maybeNull(model.StageID),
		maybeNull(model.ActiveGoalID),
		model.Period,
		model.Expected,
		model.GoalType,
		model.ExpectedSum,
		model.Currency,
		model.ExpectedType,
		maybeNull(model.CreatedByUserID),
		maybeNull(model.WorkflowID),
		model.MasterExpected,
		model.Delivered,
		model.DeliveredSum,
		model.PeriodStart,
		model.PeriodEnd,
		model.ID,
	)
	return err
}

func deleteGoal(model Goal) error {
	_, err := db.Exec(`
		UPDATE
			goals
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertNoteField(model *NoteField) error {
	row := db.QueryRow(`
		INSERT INTO note_fields(
			company_id,
			name,
			key,
			field_type,
			edit_flag,
			mandatory_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			current_timestamp
		)
		RETURNING
			id,
			created_at
	`,
		model.CompanyID,
		model.Name,
		model.Key,
		model.FieldType,
		model.EditFlag,
		model.MandatoryFlag,
	)
	return row.Scan(
		&model.ID,
		&model.CreatedAt,
	)
}

func updateNoteField(model NoteField) error {
	_, err := db.Exec(`
		UPDATE
			note_fields
		SET
			name = $1,
			key = $2,
			field_type = $3,
			edit_flag = $4,
			mandatory_flag = $5,
			updated_at = current_timestamp
		WHERE
			id = $6
	`,
		model.Name,
		model.Key,
		model.FieldType,
		model.EditFlag,
		model.MandatoryFlag,
		model.ID,
	)
	return err
}

func deleteNoteField(model NoteField) error {
	_, err := db.Exec(`
		UPDATE
			note_fields
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertOrganizationRelationship(model *OrganizationRelationship) error {
	row := db.QueryRow(`
		INSERT INTO organization_relationships(
		    company_id,
			type,
			rel_owner_org_id,
			rel_linked_org_id,
			calculated_type,
			calculated_related_org_id,
			created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			current_timestamp
		)
		RETURNING id
	`,
		model.CompanyID,
		model.Type,
		maybeNull(model.RelOwnerOrgID),
		maybeNull(model.RelLinkedOrgID),
		model.CalculatedType,
		maybeNull(model.CalculatedRelatedOrgID),
	)
	return row.Scan(&model.ID)
}

func updateOrganizationRelationship(model OrganizationRelationship) error {
	_, err := db.Exec(`
		UPDATE
			organization_relationships
		SET
			type = $1,
			rel_owner_org_id = $2,
			rel_linked_org_id = $3,
			calculated_type = $4,
			calculated_related_org_id = $5,
			updated_at = current_timestamp
		WHERE
			id = $6
	`,
		model.Type,
		maybeNull(model.RelOwnerOrgID),
		maybeNull(model.RelLinkedOrgID),
		model.CalculatedType,
		maybeNull(model.CalculatedRelatedOrgID),
		model.ID,
	)
	return err
}

func deleteOrganizationRelationship(model OrganizationRelationship) error {
	_, err := db.Exec(`
		UPDATE
			organization_relationships
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertProduct(model *Product) error {
	row := db.QueryRow(`
		INSERT INTO products(
			company_id,
			name,
			code,
			unit,
			tax,
			selectable,
			first_char,
			visible_to,
			owner_id,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			current_timestamp
		)
		RETURNING id
	`,
		model.CompanyID,
		model.Name,
		model.Code,
		model.Unit,
		model.Tax,
		model.Selectable,
		model.FirstChar,
		model.VisibleTo,
		model.OwnerID,
	)
	return row.Scan(&model.ID)
}

func updateProduct(model Product) error {
	_, err := db.Exec(`
		UPDATE
			products
		SET
			name = $1,
			code = $2,
			unit = $3,
			tax = $4,
			selectable = $5,
			first_char = $6,
			visible_to = $7,
			owner_id = $8,
			updated_at = current_timestamp
		WHERE
			id = $9
	`,
		model.Name,
		model.Code,
		model.Unit,
		model.Tax,
		model.Selectable,
		model.FirstChar,
		model.VisibleTo,
		model.OwnerID,
		model.ID,
	)
	return err
}

func deleteProduct(model Product) error {
	_, err := db.Exec(`
		UPDATE
			products
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertPrice(model *Price) error {
	row := db.QueryRow(`
		INSERT INTO prices(
			product_id,
			price,
			currency,
			cost,
			overhead_cost,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			current_timestamp
		)
		RETURNING id
	`,
		model.ProductID,
		model.Price,
		model.Currency,
		model.Cost,
		model.OverheadCost,
	)
	return row.Scan(&model.ID)
}

func updatePrice(model Price) error {
	_, err := db.Exec(`
		UPDATE
			prices
		SET
			product_id = $1,
			price = $2,
			currency = $3,
			cost = $4,
			overhead_cost = $5,
			updated_at = current_timestamp
		WHERE
			id = $6
	`,
		model.ProductID,
		model.Price,
		model.Currency,
		model.Cost,
		model.OverheadCost,
		model.ID,
	)
	return err
}

func deletePrice(model Price) error {
	_, err := db.Exec(`
		UPDATE
			prices
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertProductField(model *ProductField) error {
	row := db.QueryRow(`
		INSERT INTO product_fields(
		   	company_id,
		   	name,
			key,
			order_nr,
			picklist_data,
			field_type,
			edit_flag,
			index_visible_flag,
			details_visible_flag,
			add_visible_flag,
			important_flag,
			bulk_edit_allowed,
			use_field,
			link,
			mandatory_flag,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			current_timestamp
		)
		RETURNING id
	`,
		model.CompanyID,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.UseField,
		model.Link,
		model.MandatoryFlag,
	)
	return row.Scan(&model.ID)
}

func updateProductField(model ProductField) error {
	_, err := db.Exec(`
		UPDATE
			product_fields
		SET
		   	name = $1,
			key = $2,
			order_nr = $3,
			picklist_data = $4,
			field_type = $5,
			edit_flag = $6,
			index_visible_flag = $7,
			details_visible_flag = $8,
			add_visible_flag = $9,
			important_flag = $10,
			bulk_edit_allowed = $11,
			use_field = $12,
			link = $13,
			mandatory_flag = $14,
			updated_at = current_timestamp
		WHERE
			id = $15
	`,
		model.Name,
		model.Key,
		model.OrderNr,
		model.PicklistData,
		model.FieldType,
		model.EditFlag,
		model.IndexVisibleFlag,
		model.DetailsVisibleFlag,
		model.AddVisibleFlag,
		model.ImportantFlag,
		model.BulkEditAllowed,
		model.UseField,
		model.Link,
		model.MandatoryFlag,
		model.ID,
	)
	return err
}

func deleteProductField(model ProductField) error {
	_, err := db.Exec(`
		UPDATE
			product_fields
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func undelete(model DeletedObject) error {
	if model.Type != "workflows" &&
		model.Type != "time_entries" &&
		model.Type != "users" &&
		model.Type != "tasks" &&
		model.Type != "task_fields" &&
		model.Type != "stages" &&
		model.Type != "push_notifications" &&
		model.Type != "products" &&
		model.Type != "prices" &&
		model.Type != "persons" &&
		model.Type != "person_fields" &&
		model.Type != "organizations" &&
		model.Type != "organization_fields" &&
		model.Type != "notes" &&
		model.Type != "note_fields" &&
		model.Type != "goals" &&
		model.Type != "filters" &&
		model.Type != "files" &&
		model.Type != "currencies" &&
		model.Type != "contacts" &&
		model.Type != "company_users" &&
		model.Type != "companies" &&
		model.Type != "categories" &&
		model.Type != "activity_types" &&
		model.Type != "activity_fields" &&
		model.Type != "activities" {
		return errors.New("Invalid type")
	}

	_, err := db.Exec(`
		UPDATE
			`+model.Type+`
		SET
			deleted_at = null
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func insertTimeline(model *Timeline) error {
	row := db.QueryRow(`
		INSERT INTO timeline(
		   	user_id,
		   	under_company_id,
		   	name,
			activity_id,
			activity_field_id,
			activity_type_id,
			category_id,
			company_id,
			company_user_id,
			contact_id,
			currency_id,
			file_id,
			filter_id,
			goal_id,
			note_field_id,
			note_id,
			organization_field_id,
			organization_relationship_id,
			organization_id,
			person_id,
			person_field_id,
			price_id,
			product_field_id,
			product_id,
			push_notification_id,
			stage_id,
			task_field_id,
			task_id,
			workflow_id,
			time_entry_id,
			action,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			$19,
			$20,
			$21,
			$22,
			$23,
			$24,
			$25,
			$26,
			$27,
			$28,
			$29,
			$30,
			$31,
			current_timestamp
		)
		RETURNING id
	`,
		model.UserID,
		model.UnderCompanyID,
		model.Name,
		maybeNull(model.ActivityID),
		maybeNull(model.ActivityFieldID),
		maybeNull(model.ActivityTypeID),
		maybeNull(model.CategoryID),
		maybeNull(model.CompanyID),
		maybeNull(model.CompanyUserID),
		maybeNull(model.ContactID),
		maybeNull(model.CurrencyID),
		maybeNull(model.FileID),
		maybeNull(model.FilterID),
		maybeNull(model.GoalID),
		maybeNull(model.NoteFieldID),
		maybeNull(model.NoteID),
		maybeNull(model.OrganizationFieldID),
		maybeNull(model.OrganizationRelationshipID),
		maybeNull(model.OrganizationID),
		maybeNull(model.PersonID),
		maybeNull(model.PersonFieldID),
		maybeNull(model.PriceID),
		maybeNull(model.ProductFieldID),
		maybeNull(model.ProductID),
		maybeNull(model.PushNotificationID),
		maybeNull(model.StageID),
		maybeNull(model.TaskFieldID),
		maybeNull(model.TaskID),
		maybeNull(model.WorkflowID),
		maybeNull(model.TimeEntryID),
		model.Action,
	)
	return row.Scan(&model.ID)
}

func maybeNull(ID string) *string {
	if ID == "" {
		return nil
	}
	return &ID
}

func insertPushNotification(model *PushNotification) error {
	row := db.QueryRow(`
		INSERT INTO push_notifications(
		   	company_id,
		   	user_id,
			subscription_url,
			type,
			event,
			http_auth_user,
			http_auth_password,
			http_last_response_code,
		    created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			current_timestamp
		)
		RETURNING id
	`,
		model.CompanyID,
		model.UserID,
		model.SubscriptionURL,
		model.Type,
		model.Event,
		model.HTTPAuthUser,
		model.HTTPAuthPassword,
		model.HTTPLastResponseCode,
	)
	return row.Scan(&model.ID)
}

func updatePushNotification(model PushNotification) error {
	_, err := db.Exec(`
		UPDATE
			push_notifications
		SET
		   	user_id = $1,
			subscription_url = $2,
			type = $3,
			event = $4,
			http_auth_user = $5,
			http_auth_password = $6,
			http_last_response_code = $7,
			updated_at = current_timestamp
		WHERE
			id = $8
	`,
		model.UserID,
		model.SubscriptionURL,
		model.Type,
		model.Event,
		model.HTTPAuthUser,
		model.HTTPAuthPassword,
		model.HTTPLastResponseCode,
		model.ID,
	)
	return err
}

func deletePushNotification(model PushNotification) error {
	_, err := db.Exec(`
		UPDATE
			push_notifications
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		model.ID,
	)
	return err
}

func selectCompaniesByUser(userID string) ([]Company, error) {
	rows, err := db.Query(`
		SELECT
			id,
			name,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			companies
		WHERE
			deleted_at IS NULL
		AND
			id in (
				select company_id
				from company_users
				where user_id = $1
				and deleted_at is null
			)
		ORDER BY
			name
	`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Company
	for rows.Next() {
		var model Company

		err = rows.Scan(
			&model.ID,
			&model.Name,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectCompanyByID(companyID string) (*Company, error) {
	var model Company
	err := db.QueryRow(`
		select
			id,
			name,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			companies
		WHERE
			id = $1
	`,
		companyID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)
	return &model, err
}

func companyBelongsToUser(companyID, userID string) (bool, error) {
	var result bool
	err := db.QueryRow(`
		select
			count(1) > 0
		from
			company_users
		where
			deleted_at is null
		and
			user_id = $1
		and
			company_id = $2
	`,
		userID,
		companyID,
	).Scan(
		&result,
	)
	return result, err
}

func selectCompanyUsersByCompany(companyID string) ([]CompanyUser, error) {
	rows, err := db.Query(`
		SELECT
			company_users.id,
			company_users.user_id,
			company_users.is_admin,
			company_users.company_id,
		    company_users.created_at,
		    company_users.updated_at,
		    company_users.deleted_at,
		    users.email,
		    users.name
		FROM
			company_users
		LEFT OUTER JOIN
			users ON users.id = company_users.user_id
		WHERE
			company_users.deleted_at IS NULL
		AND
			company_users.company_id = $1
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}

	return scanCompanyUsers(rows)
}

func scanCompanyUsers(rows *sql.Rows) ([]CompanyUser, error) {
	defer rows.Close()

	var result []CompanyUser

	for rows.Next() {
		var model CompanyUser

		var name sql.NullString

		if err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.IsAdmin,
			&model.CompanyID,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
			&model.Email,
			&name,
		); err != nil {
			return nil, err
		}

		model.Name = name.String

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectTaskByID(TaskID string) (*Task, error) {
	var model Task

	var personID sql.NullString
	var personName sql.NullString
	var orgID sql.NullString
	var orgName sql.NullString
	var ownerName sql.NullString
	var nextActivityID sql.NullString

	err := db.QueryRow(`
		SELECT
		   	tasks.id,
		   	tasks.name,
			tasks.creator_user_id,
			tasks.user_id,
			tasks.person_id,
			tasks.org_id,
			tasks.stage_id,
			tasks.value,
			tasks.currency,
			tasks.stage_change_time,
			tasks.status,
			tasks.lost_reason,
			tasks.visible_to,
			tasks.close_time,
			tasks.workflow_id,
			tasks.won_time,
			tasks.first_won_time,
			tasks.lost_time,
			tasks.expected_close_date,
			tasks.stage_order_nr,
			tasks.formatted_value,
			tasks.rotten_time,
			tasks.weighted_value,
			tasks.formatted_weighted_value,
			tasks.cc_email,
			tasks.org_hidden,
			tasks.person_hidden,
		    tasks.created_at,
		    tasks.updated_at,
		    tasks.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		   	persons.name as person_name,
		   	organizations.name as org_name,
		   	stages.name as stage_name,
		   	(
		   		select activities.due_date
		   		from activities
		   		where activities.org_id = organizations.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_date,
		   	(
		   		select activities.id
		   		from activities
		   		where activities.task_id = tasks.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_id,
		   	tasks.company_id
		FROM
			tasks
		LEFT OUTER JOIN
			stages ON stages.id = tasks.stage_id
		LEFT OUTER JOIN
			users ON users.id = tasks.user_id
		LEFT OUTER JOIN
			persons ON persons.id = tasks.person_id
		LEFT OUTER JOIN
			organizations ON organizations.id = tasks.org_id
		WHERE
			tasks.id = $1
	`,
		TaskID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.CreatorUserID,
		&model.UserID,
		&personID,
		&orgID,
		&model.StageID,
		&model.Value,
		&model.Currency,
		&model.StageChangeTime,
		&model.Status,
		&model.LostReason,
		&model.VisibleTo,
		&model.CloseTime,
		&model.WorkflowID,
		&model.WonTime,
		&model.FirstWonTime,
		&model.LostTime,
		&model.ExpectedCloseDate,
		&model.StageOrderNr,
		&model.FormattedValue,
		&model.RottenTime,
		&model.WeightedValue,
		&model.FormattedWeightedValue,
		&model.CCEmail,
		&model.OrgHidden,
		&model.PersonHidden,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&ownerName,
		&personName,
		&orgName,
		&model.StageName,
		&model.NextActivityDate,
		&nextActivityID,
		&model.CompanyID,
	)
	if err != nil {
		return nil, err
	}

	model.PersonID = personID.String
	model.PersonName = personName.String
	model.OrgID = orgID.String
	model.OrgName = orgName.String
	model.OwnerName = ownerName.String
	model.NextActivityID = nextActivityID.String

	return &model, nil
}

func selectCompanyUserByID(ID string) (*CompanyUser, error) {
	var model CompanyUser

	err := db.QueryRow(`
		SELECT
			company_users.id,
			company_users.user_id,
			company_users.is_admin,
			company_users.company_id,
		    company_users.created_at,
		    company_users.updated_at,
		    company_users.deleted_at,
		    users.name,
		    users.email
		FROM
			company_users
		LEFT OUTER JOIN
			users ON users.id = company_users.user_id
		WHERE
			company_users.deleted_at IS NULL
		AND
			users.deleted_at IS NULL
		AND
			company_users.id = $1
		LIMIT 1
    `,
		ID,
	).Scan(
		&model.ID,
		&model.UserID,
		&model.IsAdmin,
		&model.CompanyID,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&model.Name,
		&model.Email,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	return &model, nil
}

func selectTimelineByCompany(companyID string) ([]Timeline, error) {
	rows, err := db.Query(`
		SELECT
			timeline.id,
		   	timeline.user_id,
		   	timeline.under_company_id,
		   	timeline.name,
			timeline.activity_id,
			timeline.activity_field_id,
			timeline.activity_type_id,
			timeline.category_id,
			timeline.company_id,
			timeline.company_user_id,
			timeline.contact_id,
			timeline.currency_id,
			timeline.file_id,
			timeline.filter_id,
			timeline.goal_id,
			timeline.note_field_id,
			timeline.note_id,
			timeline.organization_field_id,
			timeline.organization_relationship_id,
			timeline.organization_id,
			timeline.person_id,
			timeline.person_field_id,
			timeline.price_id,
			timeline.product_field_id,
			timeline.product_id,
			timeline.push_notification_id,
			timeline.stage_id,
			timeline.task_field_id,
			timeline.task_id,
			timeline.workflow_id,
			timeline.time_entry_id,
			timeline.action,
		    timeline.created_at,
		    (case when users.name = '' then users.email else users.name end) as user_name
		FROM
			timeline
		LEFT OUTER JOIN
			users ON users.id = timeline.user_id
		WHERE
			timeline.under_company_id = $1
		ORDER BY
			timeline.created_at DESC
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Timeline
	for rows.Next() {
		var model Timeline

		var activityID sql.NullString
		var activityFieldID sql.NullString
		var activityTypeID sql.NullString
		var categoryID sql.NullString
		var companyID sql.NullString
		var companyUserID sql.NullString
		var contactID sql.NullString
		var currencyID sql.NullString
		var fileID sql.NullString
		var filterID sql.NullString
		var goalID sql.NullString
		var noteFieldID sql.NullString
		var noteID sql.NullString
		var organizationFieldID sql.NullString
		var organizationRelationshipID sql.NullString
		var organizationID sql.NullString
		var personID sql.NullString
		var personFieldID sql.NullString
		var priceID sql.NullString
		var productFieldID sql.NullString
		var productID sql.NullString
		var pushNotificationID sql.NullString
		var stageID sql.NullString
		var taskFieldID sql.NullString
		var taskID sql.NullString
		var workflowID sql.NullString
		var timeEntryID sql.NullString
		var userName sql.NullString

		err = rows.Scan(
			&model.ID,
			&model.UserID,
			&model.UnderCompanyID,
			&model.Name,
			&activityID,
			&activityFieldID,
			&activityTypeID,
			&categoryID,
			&companyID,
			&companyUserID,
			&contactID,
			&currencyID,
			&fileID,
			&filterID,
			&goalID,
			&noteFieldID,
			&noteID,
			&organizationFieldID,
			&organizationRelationshipID,
			&organizationID,
			&personID,
			&personFieldID,
			&priceID,
			&productFieldID,
			&productID,
			&pushNotificationID,
			&stageID,
			&taskFieldID,
			&taskID,
			&workflowID,
			&timeEntryID,
			&model.Action,
			&model.CreatedAt,
			&userName,
		)
		if err != nil {
			return nil, err
		}

		model.ActivityID = activityID.String
		model.ActivityFieldID = activityFieldID.String
		model.ActivityTypeID = activityTypeID.String
		model.CategoryID = categoryID.String
		model.CompanyID = companyID.String
		model.CompanyUserID = companyUserID.String
		model.ContactID = contactID.String
		model.CurrencyID = currencyID.String
		model.FileID = fileID.String
		model.FilterID = filterID.String
		model.GoalID = goalID.String
		model.NoteFieldID = noteFieldID.String
		model.NoteID = noteID.String
		model.OrganizationFieldID = organizationFieldID.String
		model.OrganizationRelationshipID = organizationRelationshipID.String
		model.OrganizationID = organizationID.String
		model.PersonID = personID.String
		model.PersonFieldID = personFieldID.String
		model.PriceID = priceID.String
		model.ProductFieldID = productFieldID.String
		model.ProductID = productID.String
		model.PushNotificationID = pushNotificationID.String
		model.StageID = stageID.String
		model.TaskFieldID = taskFieldID.String
		model.TaskID = taskID.String
		model.WorkflowID = workflowID.String
		model.TimeEntryID = timeEntryID.String
		model.UserName = userName.String

		result = append(result, model)
	}
	return result, rows.Err()
}

func hasWorkflows(userID string) (bool, error) {
	var result bool
	err := db.QueryRow(`
		select
			count(1) > 0
		from
			workflows
		where
			company_id in (
				select company_id
				from company_users
				where user_id = $1
			)
		and
			deleted_at is null
	`,
		userID,
	).Scan(
		&result,
	)
	return result, err
}

func hasActivityTypes(userID string) (bool, error) {
	var result bool
	err := db.QueryRow(`
		select
			count(1) > 0
		from
			activity_types
		where
			company_id in (
				select company_id
				from company_users
				where user_id = $1
			)
		and
			deleted_at is null
	`,
		userID,
	).Scan(
		&result,
	)
	return result, err
}

func hasCurrencies(userID string) (bool, error) {
	var result bool
	err := db.QueryRow(`
		select
			count(1) > 0
		from
			currencies
		where
			company_id in (
				select company_id
				from company_users
				where user_id = $1
			)
		and
			deleted_at is null
	`,
		userID,
	).Scan(
		&result,
	)
	return result, err
}

func selectActivitiesByTask(taskID string) ([]Activity, error) {
	rows, err := db.Query(`
		SELECT
		   	activities.id,
		   	activities.name,
			activities.company_id,
			activities.user_id,
			activities.done,
			activities.reference_type,
			activities.reference_id,
			activities.due_date,
			activities.duration,
			activities.marked_as_done_time,
			activities.task_id,
			activities.org_id,
			activities.person_id,
			activities.assigned_to_user_id,
			activities.created_by_user_id,
		    activities.created_at,
		    activities.updated_at,
		    activities.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name,
		    persons.name as person_name,
		    tasks.name as task_name,
		    organizations.name as org_name,
		    activities.type_id,
		    activity_types.name as type,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as person_email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as person_phone,
		   	assigned_users.name as assigned_to_user_name
		FROM
			activities
		LEFT OUTER JOIN
			users ON users.id = activities.user_id
		LEFT OUTER JOIN
			users AS assigned_users ON assigned_users.id = activities.assigned_to_user_id
		LEFT OUTER JOIN
			persons ON persons.id = activities.person_id
		LEFT OUTER JOIN
			tasks ON tasks.id = activities.task_id
		LEFT OUTER JOIN
			organizations ON organizations.id = activities.org_id
		LEFT OUTER JOIN
			activity_types ON activity_types.id = activities.type_id
		WHERE
			activities.deleted_at IS NULL
		AND
			activities.task_id = $1
	`,
		taskID,
	)
	if err != nil {
		return nil, err
	}

	return scanActivities(rows)
}

func selectActivitiesByPerson(personID string) ([]Activity, error) {
	rows, err := db.Query(`
		SELECT
		   	activities.id,
		   	activities.name,
			activities.company_id,
			activities.user_id,
			activities.done,
			activities.reference_type,
			activities.reference_id,
			activities.due_date,
			activities.duration,
			activities.marked_as_done_time,
			activities.task_id,
			activities.org_id,
			activities.person_id,
			activities.assigned_to_user_id,
			activities.created_by_user_id,
		    activities.created_at,
		    activities.updated_at,
		    activities.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name,
		    persons.name as person_name,
		    tasks.name as task_name,
		    organizations.name as org_name,
		    activities.type_id,
		    activity_types.name as type,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as person_email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as person_phone,
		   	assigned_users.name as assigned_to_user_name
		FROM
			activities
		LEFT OUTER JOIN
			users ON users.id = activities.user_id
		LEFT OUTER JOIN
			users AS assigned_users ON assigned_users.id = activities.assigned_to_user_id
		LEFT OUTER JOIN
			persons ON persons.id = activities.person_id
		LEFT OUTER JOIN
			tasks ON tasks.id = activities.task_id
		LEFT OUTER JOIN
			organizations ON organizations.id = activities.org_id
		LEFT OUTER JOIN
			activity_types ON activity_types.id = activities.type_id
		WHERE
			activities.deleted_at IS NULL
		AND
			activities.person_id = $1
	`,
		personID,
	)
	if err != nil {
		return nil, err
	}

	return scanActivities(rows)
}

func selectOrganizationByID(orgID string) (*Organization, error) {
	var model Organization

	var categoryID sql.NullString
	var ownerName sql.NullString
	var nextActivityID sql.NullString

	err := db.QueryRow(`
		SELECT
		   	organizations.id,
		   	organizations.name,
			organizations.company_id,
			organizations.owner_id,
			organizations.country_code,
			organizations.address,
			organizations.address_subpremise,
			organizations.address_street_number,
			organizations.address_route,
			organizations.address_sublocality,
			organizations.address_locality,
			organizations.address_admin_area_level_1,
			organizations.address_admin_area_level_2,
			organizations.address_country,
			organizations.address_postal_code,
			organizations.category_id,
			organizations.first_char,
			organizations.visible_to,
		    organizations.created_at,
		    organizations.updated_at,
		    organizations.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.org_id = organizations.id
		    	and deleted_at is null
		    	and (won_time is null and lost_time is null)
		    ) as open_tasks_count,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.org_id = organizations.id
		    	and deleted_at is null
		    	and (won_time is not null or lost_time is not null)
		    ) as closed_tasks_count,
		    (
		    	select count(1)
		    	from persons
		    	where persons.org_id = organizations.id
		    	and persons.deleted_at is null
		    ) as people_count,
		    (
		    	select min(due_date)
		    	from activities
		    	where deleted_at is null
		    	and activities.org_id = organizations.id
		    ) as next_activity_date,
		    (
		    	select activities.id
		    	from activities
		    	where deleted_at is null
		    	and activities.org_id = organizations.id
		    	order by activities.due_date asc
		    	limit 1
		    ) as next_activity_id
		FROM
			organizations
		LEFT OUTER JOIN
			users ON users.id = organizations.owner_id
		WHERE
			organizations.deleted_at IS NULL
		AND
			organizations.id = $1
		LIMIT 1
    `,
		orgID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.CompanyID,
		&model.OwnerID,
		&model.CountryCode,
		&model.Address,
		&model.AddressSubpremise,
		&model.AddressStreetNumber,
		&model.AddressRoute,
		&model.AddressSublocality,
		&model.AddressLocality,
		&model.AddressAdminAreaLevel1,
		&model.AddressAdminAreaLevel2,
		&model.AddressCountry,
		&model.AddressPostalCode,
		&categoryID,
		&model.FirstChar,
		&model.VisibleTo,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&ownerName,
		&model.OpenTasksCount,
		&model.ClosedTasksCount,
		&model.PeopleCount,
		&model.NextActivityDate,
		&nextActivityID,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	model.CategoryID = categoryID.String
	model.OwnerName = ownerName.String
	model.NextActivityID = nextActivityID.String

	return &model, nil
}

func selectPersonByID(personID string) (*Person, error) {
	var model Person

	var orgID sql.NullString
	var ownerName sql.NullString
	var nextActivityID sql.NullString
	var email sql.NullString
	var phone sql.NullString
	var orgName sql.NullString

	err := db.QueryRow(`
		SELECT
		   	persons.id,
		   	persons.name,
			persons.company_id,
			persons.owner_id,
			persons.org_id,
			persons.first_name,
		    persons.created_at,
		    persons.updated_at,
		    persons.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.person_id = persons.id
		    	and deleted_at is null
		    	and (won_time is null and lost_time is null)
		    ) as open_tasks_count,
		    (
		    	select count(1)
		    	from tasks
		    	where tasks.person_id = persons.id
		    	and deleted_at is null
		    	and (won_time is not null or lost_time is not null)
		    ) as closed_tasks_count,
		   	(
		   		select activities.due_date
		   		from activities
		   		where activities.person_id = persons.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_date,
		   	(
		   		select activities.id
		   		from activities
		   		where activities.person_id = persons.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_id,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as phone,
		   	organizations.name AS org_name
		FROM
			persons
		LEFT OUTER JOIN
			users ON users.id = persons.owner_id
		LEFT OUTER JOIN
			organizations ON organizations.id = persons.org_id
		WHERE
			persons.id = $1
		ORDER BY
			persons.name
    `,
		personID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.CompanyID,
		&model.OwnerID,
		&orgID,
		&model.FirstName,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&ownerName,
		&model.OpenTasksCount,
		&model.ClosedTasksCount,
		&model.NextActivityDate,
		&nextActivityID,
		&email,
		&phone,
		&orgName,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	model.OrgID = orgID.String
	model.OwnerName = ownerName.String
	model.NextActivityID = nextActivityID.String
	model.Email = email.String
	model.Phone = phone.String
	model.OrgName = orgName.String

	return &model, nil
}

func selectActivityByID(activityID string) (*Activity, error) {
	var model Activity

	var TaskID sql.NullString
	var orgID sql.NullString
	var personID sql.NullString
	var assignedToUserID sql.NullString
	var createdByUserID sql.NullString
	var userName sql.NullString
	var personName sql.NullString
	var taskName sql.NullString
	var orgName sql.NullString
	var typeID sql.NullString
	var typeName sql.NullString
	var personEmail sql.NullString
	var personPhone sql.NullString
	var assignedToUserName sql.NullString

	err := db.QueryRow(`
		SELECT
		   	activities.id,
		   	activities.name,
			activities.company_id,
			activities.user_id,
			activities.done,
			activities.reference_type,
			activities.reference_id,
			activities.due_date,
			activities.duration,
			activities.marked_as_done_time,
			activities.task_id,
			activities.org_id,
			activities.person_id,
			activities.assigned_to_user_id,
			activities.created_by_user_id,
		    activities.created_at,
		    activities.updated_at,
		    activities.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name,
		    persons.name as person_name,
		    tasks.name as task_name,
		    organizations.name as org_name,
		    activities.type_id,
		    activity_types.name as type,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as person_email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as person_phone,
		   	assigned_users.name as assigned_to_user_name
		FROM
			activities
		LEFT OUTER JOIN
			users ON users.id = activities.user_id
		LEFT OUTER JOIN
			users AS assigned_users on assigned_users.id = activities.assigned_to_user_id
		LEFT OUTER JOIN
			persons ON persons.id = activities.person_id
		LEFT OUTER JOIN
			tasks ON tasks.id = activities.task_id
		LEFT OUTER JOIN
			organizations ON organizations.id = activities.org_id
		LEFT OUTER JOIN
			activity_types ON activity_types.id = activities.type_id
		WHERE
			activities.deleted_at IS NULL
		AND
			activities.id = $1
    `,
		activityID,
	).Scan(
		&model.ID,
		&model.Name,
		&model.CompanyID,
		&model.UserID,
		&model.Done,
		&model.ReferenceType,
		&model.ReferenceID,
		&model.DueDate,
		&model.Duration,
		&model.DoneAt,
		&TaskID,
		&orgID,
		&personID,
		&assignedToUserID,
		&createdByUserID,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
		&userName,
		&personName,
		&taskName,
		&orgName,
		&typeID,
		&typeName,
		&personEmail,
		&personPhone,
		&assignedToUserName,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}

	model.TaskID = TaskID.String
	model.OrgID = orgID.String
	model.PersonID = personID.String
	model.AssignedToUserID = assignedToUserID.String
	model.CreatedByUserID = createdByUserID.String
	model.OwnerName = userName.String
	model.PersonName = personName.String
	model.TaskTitle = taskName.String
	model.OrgName = orgName.String
	model.TypeID = typeID.String
	model.Type = typeName.String
	model.PersonEmail = personEmail.String
	model.PersonPhone = personPhone.String
	model.AssignedToUserName = assignedToUserName.String

	return &model, nil
}

func scanActivities(rows *sql.Rows) ([]Activity, error) {
	defer rows.Close()
	var result []Activity
	for rows.Next() {
		var model Activity

		var taskID sql.NullString
		var orgID sql.NullString
		var personID sql.NullString
		var assignedToUserID sql.NullString
		var createdByUserID sql.NullString
		var userName sql.NullString
		var personName sql.NullString
		var taskName sql.NullString
		var orgName sql.NullString
		var typeID sql.NullString
		var typeName sql.NullString
		var personEmail sql.NullString
		var personPhone sql.NullString
		var assignedToUserName sql.NullString

		if err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.CompanyID,
			&model.UserID,
			&model.Done,
			&model.ReferenceType,
			&model.ReferenceID,
			&model.DueDate,
			&model.Duration,
			&model.DoneAt,
			&taskID,
			&orgID,
			&personID,
			&assignedToUserID,
			&createdByUserID,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
			&userName,
			&personName,
			&taskName,
			&orgName,
			&typeID,
			&typeName,
			&personEmail,
			&personPhone,
			&assignedToUserName,
		); err != nil {
			return nil, err
		}

		model.TaskID = taskID.String
		model.OrgID = orgID.String
		model.PersonID = personID.String
		model.AssignedToUserID = assignedToUserID.String
		model.CreatedByUserID = createdByUserID.String
		model.OwnerName = userName.String
		model.PersonName = personName.String
		model.TaskTitle = taskName.String
		model.OrgName = orgName.String
		model.TypeID = typeID.String
		model.Type = typeName.String
		model.PersonEmail = personEmail.String
		model.PersonPhone = personPhone.String
		model.AssignedToUserName = assignedToUserName.String

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectNotesByOrganization(orgID string) ([]Note, error) {
	rows, err := db.Query(`
		SELECT
		   	notes.id,
		   	notes.name,
			notes.company_id,
			notes.user_id,
			notes.task_id,
			notes.person_id,
			notes.org_id,
			notes.pinned_to_task_flag,
			notes.pinned_to_person_flag,
			notes.pinned_to_organization_flag,
			notes.last_update_user_id,
		    notes.created_at,
		    notes.updated_at,
		    notes.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name
		FROM
			notes
		LEFT OUTER JOIN
			users ON users.id = notes.user_id
		WHERE
			notes.deleted_at IS NULL
		AND
			notes.org_id = $1
	`,
		orgID,
	)
	if err != nil {
		return nil, err
	}

	return scanNotes(rows)
}

func selectTasksByOrganization(orgID string) ([]Task, error) {
	rows, err := db.Query(`
		SELECT
		   	tasks.id,
		   	tasks.name,
			tasks.creator_user_id,
			tasks.user_id,
			tasks.person_id,
			tasks.org_id,
			tasks.stage_id,
			tasks.value,
			tasks.currency,
			tasks.stage_change_time,
			tasks.status,
			tasks.lost_reason,
			tasks.visible_to,
			tasks.close_time,
			tasks.workflow_id,
			tasks.won_time,
			tasks.first_won_time,
			tasks.lost_time,
			tasks.expected_close_date,
			tasks.stage_order_nr,
			tasks.formatted_value,
			tasks.rotten_time,
			tasks.weighted_value,
			tasks.formatted_weighted_value,
			tasks.cc_email,
			tasks.org_hidden,
			tasks.person_hidden,
		    tasks.created_at,
		    tasks.updated_at,
		    tasks.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as owner_name,
		   	persons.name as person_name,
		   	organizations.name as org_name,
		   	stages.name as stage_name,
		   	(
		   		select activities.due_date
		   		from activities
		   		where activities.task_id = tasks.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_date,
		   	(
		   		select activities.id
		   		from activities
		   		where activities.task_id = tasks.id
		   		and activities.deleted_at is null
		   		order by activities.due_date asc
		   		limit 1
		   	) as next_activity_id,
		   	tasks.company_id
		FROM
			tasks
		LEFT OUTER JOIN
			users ON users.id = tasks.user_id
		LEFT OUTER JOIN
			persons ON persons.id = tasks.person_id
		LEFT OUTER JOIN
			organizations ON organizations.id = tasks.org_id
		LEFT OUTER JOIN
			stages ON stages.id = tasks.stage_id
		WHERE
			tasks.deleted_at IS NULL
		AND
			tasks.org_id = $1
	`,
		orgID,
	)
	if err != nil {
		return nil, err
	}

	return scanTasks(rows)
}

func selectActivitiesByOrganization(orgID string) ([]Activity, error) {
	rows, err := db.Query(`
		SELECT
		   	activities.id,
		   	activities.name,
			activities.company_id,
			activities.user_id,
			activities.done,
			activities.reference_type,
			activities.reference_id,
			activities.due_date,
			activities.duration,
			activities.marked_as_done_time,
			activities.task_id,
			activities.org_id,
			activities.person_id,
			activities.assigned_to_user_id,
			activities.created_by_user_id,
		    activities.created_at,
		    activities.updated_at,
		    activities.deleted_at,
		    (case when users.name = '' then users.email else users.name end) as user_name,
		    persons.name as person_name,
		    tasks.name as task_name,
		    organizations.name as org_name,
		    activities.type_id,
		    activity_types.name as type,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'email'
		   		and contacts.deleted_at is null
		   	) as person_email,
		   	(
		   		select string_agg(contacts.name, ', ')
		   		from contacts
		   		where contacts.person_id = persons.id
		   		and contacts.type = 'phone'
		   		and contacts.deleted_at is null
		   	) as person_phone,
		   	assigned_users.name as assigned_to_user_name
		FROM
			activities
		LEFT OUTER JOIN
			users ON users.id = activities.user_id
		LEFT OUTER JOIN
			users AS assigned_users ON assigned_users.id = activities.assigned_to_user_id
		LEFT OUTER JOIN
			persons ON persons.id = activities.person_id
		LEFT OUTER JOIN
			tasks ON tasks.id = activities.task_id
		LEFT OUTER JOIN
			organizations ON organizations.id = activities.org_id
		LEFT OUTER JOIN
			activity_types ON activity_types.id = activities.type_id
		WHERE
			activities.deleted_at IS NULL
		AND
			activities.org_id = $1
	`,
		orgID,
	)
	if err != nil {
		return nil, err
	}

	return scanActivities(rows)
}

func selectDeletedObjectsByCompany(companyID string) ([]DeletedObject, error) {
	rows, err := db.Query(`
		(
			SELECT
				'time_entries',
				id,
				name,
			    created_at,
			    updated_at,
			    deleted_at
			FROM
				time_entries
			WHERE
				deleted_at IS NOT NULL
			and
				company_id = $1
		)
		UNION
		(
			SELECT
				'workflows',
				id,
				name,
			    created_at,
			    updated_at,
			    deleted_at
			FROM
				workflows
			WHERE
				deleted_at IS NOT NULL
			and
				company_id = $1
		)
		UNION
		(
			SELECT
				'activities',
				id,
				name,
			    created_at,
			    updated_at,
			    deleted_at
			FROM
				activities
			WHERE
				deleted_at IS NOT NULL
			and
				company_id = $1
		)
		UNION
		(
			SELECT
				'activity_fields',
				id,
				name,
			    created_at,
			    updated_at,
			    deleted_at
			FROM
				activity_fields
			WHERE
				deleted_at IS NOT NULL
			and
				company_id = $1
		)
		UNION
		(
			SELECT
				'activity_types',
				id,
				name,
			    created_at,
			    updated_at,
			    deleted_at
			FROM
				activity_types
			WHERE
				deleted_at IS NOT NULL
			and
				company_id = $1
		)
		UNION
		(
			SELECT
				'company_users',
				company_users.id,
				users.name,
			    company_users.created_at,
			    company_users.updated_at,
			    company_users.deleted_at
			FROM
				company_users
			LEFT OUTER JOIN
				users on users.id = company_users.user_id
			WHERE
				company_users.deleted_at IS NOT NULL
			and
				company_users.company_id = $1
		)
		UNION
		(
			SELECT
				'contacts',
				contacts.id,
				contacts.name,
			    contacts.created_at,
			    contacts.updated_at,
			    contacts.deleted_at
			FROM
				contacts
			LEFT OUTER JOIN
				persons on persons.id = contacts.person_id
			WHERE
				contacts.deleted_at IS NOT NULL
			and
				persons.company_id = $1
		)
		UNION
		(
			SELECT
				'currencies',
				currencies.id,
				currencies.name,
			    currencies.created_at,
			    currencies.updated_at,
			    currencies.deleted_at
			FROM
				currencies
			WHERE
				currencies.deleted_at IS NOT NULL
			and
				currencies.company_id = $1
		)
		UNION
		(
			SELECT
				'files',
				files.id,
				files.name,
			    files.created_at,
			    files.updated_at,
			    files.deleted_at
			FROM
				files
			WHERE
				files.deleted_at IS NOT NULL
			and
				files.user_id in (
					select user_id
					from company_users
					where company_id = $1
				)
		)
		UNION
		(
			SELECT
				'filters',
				filters.id,
				filters.name,
			    filters.created_at,
			    filters.updated_at,
			    filters.deleted_at
			FROM
				filters
			WHERE
				filters.deleted_at IS NOT NULL
			and
				filters.company_id = $1
		)
		UNION
		(
			SELECT
				'goals',
				goals.id,
				goals.name,
			    goals.created_at,
			    goals.updated_at,
			    goals.deleted_at
			FROM
				goals
			WHERE
				goals.deleted_at IS NOT NULL
			and
				goals.company_id = $1
		)
		UNION
		(
			SELECT
				'note_fields',
				note_fields.id,
				note_fields.name,
			    note_fields.created_at,
			    note_fields.updated_at,
			    note_fields.deleted_at
			FROM
				note_fields
			WHERE
				note_fields.deleted_at IS NOT NULL
			and
				note_fields.company_id = $1
		)
		UNION
		(
			SELECT
				'notes',
				notes.id,
				notes.name,
			    notes.created_at,
			    notes.updated_at,
			    notes.deleted_at
			FROM
				notes
			WHERE
				notes.deleted_at IS NOT NULL
			and
				notes.company_id = $1
		)		
		UNION
		(
			SELECT
				'organization_fields',
				organization_fields.id,
				organization_fields.name,
			    organization_fields.created_at,
			    organization_fields.updated_at,
			    organization_fields.deleted_at
			FROM
				organization_fields
			WHERE
				organization_fields.deleted_at IS NOT NULL
			and
				organization_fields.company_id = $1
		)		
		UNION
		(
			SELECT
				'organizations',
				organizations.id,
				organizations.name,
			    organizations.created_at,
			    organizations.updated_at,
			    organizations.deleted_at
			FROM
				organizations
			WHERE
				organizations.deleted_at IS NOT NULL
			and
				organizations.company_id = $1
		)		
		UNION
		(
			SELECT
				'person_fields',
				person_fields.id,
				person_fields.name,
			    person_fields.created_at,
			    person_fields.updated_at,
			    person_fields.deleted_at
			FROM
				person_fields
			WHERE
				person_fields.deleted_at IS NOT NULL
			and
				person_fields.company_id = $1
		)		
		UNION
		(
			SELECT
				'persons',
				persons.id,
				persons.name,
			    persons.created_at,
			    persons.updated_at,
			    persons.deleted_at
			FROM
				persons
			WHERE
				persons.deleted_at IS NOT NULL
			and
				persons.company_id = $1
		)		
		UNION
		(
			SELECT
				'products',
				products.id,
				products.name,
			    products.created_at,
			    products.updated_at,
			    products.deleted_at
			FROM
				products
			WHERE
				products.deleted_at IS NOT NULL
			and
				products.company_id = $1
		)		
		UNION
		(
			SELECT
				'product_fields',
				product_fields.id,
				product_fields.name,
			    product_fields.created_at,
			    product_fields.updated_at,
			    product_fields.deleted_at
			FROM
				product_fields
			WHERE
				product_fields.deleted_at IS NOT NULL
			and
				product_fields.company_id = $1
		)		
		UNION
		(
			SELECT
				'push_notifications',
				push_notifications.id,
				push_notifications.subscription_url,
			    push_notifications.created_at,
			    push_notifications.updated_at,
			    push_notifications.deleted_at
			FROM
				push_notifications
			WHERE
				push_notifications.deleted_at IS NOT NULL
			and
				push_notifications.company_id = $1
		)		
		UNION
		(
			SELECT
				'stages',
				stages.id,
				stages.name,
			    stages.created_at,
			    stages.updated_at,
			    stages.deleted_at
			FROM
				stages
			WHERE
				stages.deleted_at IS NOT NULL
			and
				stages.workflow_id in (
					select id from workflows
					where company_id = $1
				)
		)		
		UNION
		(
			SELECT
				'task_fields',
				task_fields.id,
				task_fields.name,
			    task_fields.created_at,
			    task_fields.updated_at,
			    task_fields.deleted_at
			FROM
				task_fields
			WHERE
				task_fields.deleted_at IS NOT NULL
			and
				task_fields.company_id = $1
		)		
		UNION
		(
			SELECT
				'tasks',
				tasks.id,
				tasks.name,
			    tasks.created_at,
			    tasks.updated_at,
			    tasks.deleted_at
			FROM
				tasks
			WHERE
				tasks.deleted_at IS NOT NULL
			and
				tasks.company_id = $1
		)		
		UNION
		(
			SELECT
				'prices',
				prices.id,
				products.name,
			    prices.created_at,
			    prices.updated_at,
			    prices.deleted_at
			FROM
				prices
			LEFT OUTER JOIN 
				products on products.id = prices.product_id
			WHERE
				prices.deleted_at IS NOT NULL
			and
				products.company_id = $1
		)		
		ORDER BY
			deleted_at DESC
	`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []DeletedObject
	for rows.Next() {
		var model DeletedObject

		err = rows.Scan(
			&model.Type,
			&model.ID,
			&model.Name,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, model)
	}
	return result, rows.Err()
}

func stopRunningTimeEntries(userID string) error {
	_, err := db.Exec(`
		UPDATE
			time_entries
		SET
			finished_at = current_timestamp
		WHERE
			user_id = $1
		AND
			finished_at IS NULL
		AND
			deleted_at IS NULL
	`,
		userID,
	)
	return err
}

func insertTimeEntry(model *TimeEntry) error {
	if nil == model.FinishedAt {
		if err := stopRunningTimeEntries(model.UserID); err != nil {
			return err
		}
	}

	row := db.QueryRow(`
		INSERT INTO time_entries(
			user_id,
			company_id,
			task_id,
			activity_id,
			started_at,
			finished_at,
			name,
			created_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			current_timestamp
		)
		RETURNING id
	`,
		model.UserID,
		model.CompanyID,
		maybeNull(model.TaskID),
		maybeNull(model.ActivityID),
		model.StartedAt,
		model.FinishedAt,
		model.Name,
	)
	return row.Scan(&model.ID)
}

func updateTimeEntry(model TimeEntry) error {
	_, err := db.Exec(`
		UPDATE
			time_entries
		SET
			task_id = $1,
			activity_id = $2,
			started_at = $3,
			finished_at = $4,
			name = $5,
			updated_at = current_timestamp
		WHERE
			id = $6
	`,
		maybeNull(model.TaskID),
		maybeNull(model.ActivityID),
		model.StartedAt,
		model.FinishedAt,
		model.Name,
		model.ID,
	)
	return err
}

func deleteTimeEntry(timeEntryID string) error {
	_, err := db.Exec(`
		UPDATE
			time_entries
		SET
			deleted_at = current_timestamp
		WHERE
			id = $1
	`,
		timeEntryID,
	)
	return err
}

const selectTimeEntriesByUserAndCompanySQL = `
		(
			SELECT
				time_entries.id,
				time_entries.user_id,
				time_entries.company_id,
				time_entries.name,
				time_entries.started_at,
				time_entries.finished_at,
				time_entries.task_id,
				tasks.name as task_name,
				time_entries.activity_id,
				activities.name as activity_name,
			    time_entries.created_at,
			    time_entries.updated_at,
			    time_entries.deleted_at
			FROM
				time_entries
			LEFT OUTER JOIN
				tasks on tasks.id = time_entries.task_id
			LEFT OUTER JOIN
				activities on activities.id = time_entries.activity_id
			WHERE
				time_entries.deleted_at IS NULL
			AND
				time_entries.user_id = $1
			AND
				time_entries.company_id = $2
			AND
				time_entries.finished_at IS NULL
		)

		UNION

		(
			SELECT
				time_entries.id,
				time_entries.user_id,
				time_entries.company_id,
				time_entries.name,
				time_entries.started_at,
				time_entries.finished_at,
				time_entries.task_id,
				tasks.name as task_name,
				time_entries.activity_id,
				activities.name as activity_name,
			    time_entries.created_at,
			    time_entries.updated_at,
			    time_entries.deleted_at
			FROM
				time_entries
			LEFT OUTER JOIN
				tasks on tasks.id = time_entries.task_id
			LEFT OUTER JOIN
				activities on activities.id = time_entries.activity_id
			WHERE
				time_entries.deleted_at IS NULL
			AND
				time_entries.user_id = $1
			AND
				time_entries.company_id = $2
			AND
				time_entries.finished_at IS NOT NULL
			%s
		)

		ORDER BY
			started_at
	`

func selectTimeEntriesByUserAndCompany(userID, companyID string, fromTime, untilTime *time.Time) ([]TimeEntry, error) {
	var rows *sql.Rows
	var err error
	if fromTime != nil && untilTime != nil {
		query := fmt.Sprintf(selectTimeEntriesByUserAndCompanySQL, `
			AND time_entries.started_at >= $3 AND time_entries.started_at < $4`)
		rows, err = db.Query(query, userID, companyID, fromTime, untilTime)
	} else {
		query := fmt.Sprintf(selectTimeEntriesByUserAndCompanySQL, ``)
		rows, err = db.Query(query, userID, companyID)
	}

	if err != nil {
		return nil, err
	}

	return scanTimeEntries(rows)
}

func scanTimeEntries(rows *sql.Rows) ([]TimeEntry, error) {
	defer rows.Close()

	var result []TimeEntry

	for rows.Next() {
		var model TimeEntry

		var taskID sql.NullString
		var taskName sql.NullString
		var activityID sql.NullString
		var activityName sql.NullString

		if err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.CompanyID,
			&model.Name,
			&model.StartedAt,
			&model.FinishedAt,
			&taskID,
			&taskName,
			&activityID,
			&activityName,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		); err != nil {
			return nil, err
		}

		model.TaskID = taskID.String
		model.TaskName = taskName.String
		model.ActivityID = activityID.String
		model.ActivityName = activityName.String

		result = append(result, model)
	}
	return result, rows.Err()
}

func selectTimeEntryByID(ID string) (*TimeEntry, error) {
	var model TimeEntry

	var taskID sql.NullString
	var taskName sql.NullString
	var activityID sql.NullString
	var activityName sql.NullString

	err := db.QueryRow(`
		SELECT
			time_entries.id,
			time_entries.user_id,
			time_entries.company_id,
			time_entries.name,
			time_entries.started_at,
			time_entries.finished_at,
			time_entries.task_id,
			tasks.name as task_name,
			time_entries.activity_id,
			activities.name as activity_name,
		    time_entries.created_at,
		    time_entries.updated_at,
		    time_entries.deleted_at
		FROM
			time_entries
		LEFT OUTER JOIN
			tasks on tasks.id = time_entries.task_id
		LEFT OUTER JOIN
			activities on activities.id = time_entries.activity_id
		WHERE
			time_entries.id = $1
	`,
		ID,
	).Scan(
		&model.ID,
		&model.UserID,
		&model.CompanyID,
		&model.Name,
		&model.StartedAt,
		&model.FinishedAt,
		&taskID,
		&taskName,
		&activityID,
		&activityName,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)

	model.TaskID = taskID.String
	model.TaskName = taskName.String
	model.ActivityID = activityID.String
	model.ActivityName = activityName.String

	return &model, err
}

func selectUserIDByApiToken(token string) (string, error) {
	var result string
	err := db.QueryRow(`
		select
			id
		from
			users
		where
			api_token is not null
		and
			api_token <> ''
		and
			api_token = $1
		limit 1
	`,
		token,
	).Scan(
		&result,
	)
	return result, err
}

func insertUserEvent(model *UserEvent) error {
	row := db.QueryRow(`
		INSERT INTO user_events(
			name,
			user_id,
			created_at
		)
		VALUES(
			$1,
			$2,
			current_timestamp
		)
		RETURNING id
	`,
		model.Name,
		model.UserID,
	)
	return row.Scan(&model.ID)
}

func selectUserEventsByUser(userID string) ([]UserEvent, error) {
	rows, err := db.Query(`
		SELECT
			id,
			user_id,
			name,
		    created_at,
		    updated_at,
		    deleted_at
		FROM
			user_events
		WHERE
			deleted_at IS NULL
		AND
			user_id = $1
	`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return scanUserEvents(rows)
}

func scanUserEvents(rows *sql.Rows) ([]UserEvent, error) {
	defer rows.Close()

	var result []UserEvent

	for rows.Next() {
		var model UserEvent

		if err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		); err != nil {
			return nil, err
		}

		result = append(result, model)
	}
	return result, rows.Err()
}
