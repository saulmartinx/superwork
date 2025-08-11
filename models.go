package main

import (
	"fmt"
	"strings"
	"time"
)

// Base is a base model
type Base struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Activation struct {
	Base
	UserID string `json:"user_id"`
}

// User is a user in the backend.
// Can be an admin-user, too
type User struct {
	Base
	Email              string     `json:"email"`
	PasswordHash       string     `json:"-"`
	Password           string     `json:"password,omitempty"`
	YearOfBirth        *int       `json:"year_of_birth,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	Picture            *string    `json:"picture,omitempty"`
	DefaultCurrency    string     `json:"default_currency,omitempty"`
	Locale             string     `json:"locale"`
	Lang               int        `json:"lang"`
	Activated          bool       `json:"activated"`
	LastLogin          *time.Time `json:"last_login"`
	RoleID             string     `json:"role_id,omitempty"`
	TimezoneName       string     `json:"timezone_name"`
	IconURL            string     `json:"icon_url"`
	ActiveCompanyID    string     `json:"active_company_id"`
	ActiveWorkflowID   string     `json:"active_workflow_id"`
	HasPassword        bool       `json:"has_password"`
	ActiveCompanyName  string     `json:"active_company_name"`
	ActiveWorkflowName string     `json:"active_workflow_name"`
}

type CompanyUser struct {
	Base
	UserID    string `json:"user_id"`
	IsAdmin   bool   `json:"is_admin"`
	CompanyID string `json:"company_id"`
	Email     string `json:"email"`
}

type DeletedObject struct {
	Base
	Type string `json:"type"`
}

type stats struct {
	Users         int64 `json:"users"`
	Companies     int64 `json:"companies"`
	Workflows     int64 `json:"workflows"`
	Organizations int64 `json:"organizations"`
	People        int64 `json:"people"`
	Tasks         int64 `json:"tasks"`
	Activities    int64 `json:"activities"`
	TimeEntries   int64 `json:"time_entries"`
}

type Company struct {
	Base
}

type TimeEntry struct {
	Base
	UserID     string     `json:"user_id"`
	CompanyID  string     `json:"company_id"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	TaskID     string     `json:"task_id"`
	ActivityID string     `json:"activity_id"`

	TaskName     string `json:"task_name"`
	ActivityName string `json:"activity_name"`
}

func (model TimeEntry) Duration() time.Duration {
	return model.FinishedAt.Sub(model.StartedAt)
}

func (model TimeEntry) DurationSeconds() float64 {
	return model.Duration().Seconds()
}

func (model TimeEntry) DurationString() string {
	totalSeconds := int64(model.Duration().Seconds())

	seconds := totalSeconds % 60
	totalMinutes := totalSeconds / 60
	minutes := totalMinutes % 60
	hours := totalMinutes / 60

	return fmt.Sprintf("%.2d:%.2d:%.2d.00", hours, minutes, seconds)
}

type Activity struct {
	Base
	CompanyID        string     `json:"company_id"`
	UserID           string     `json:"user_id"`
	Done             bool       `json:"done"`
	TypeID           string     `json:"type_id"`
	ReferenceType    string     `json:"reference_type"`
	ReferenceID      string     `json:"reference_id"`
	DueDate          *time.Time `json:"due_date"`
	Duration         string     `json:"duration"`
	DoneAt           *time.Time `json:"marked_as_done_time"`
	TaskID           string     `json:"task_id"`
	OrgID            string     `json:"org_id"`
	PersonID         string     `json:"person_id"`
	Note             string     `json:"note"`
	AssignedToUserID string     `json:"assigned_to_user_id"`
	CreatedByUserID  string     `json:"created_by_user_id"`

	Type               string `json:"type"`
	OwnerName          string `json:"owner_name"`
	AssignedToUserName string `json:"assigned_to_user_name"`
	TaskTitle          string `json:"task_title"`
	OrgName            string `json:"org_name"`
	PersonName         string `json:"person_name"`
	PersonEmail        string `json:"person_email"`
	PersonPhone        string `json:"person_phone"`
}

type ModelWithPerson interface {
	GetPersonName() string
	GetPersonID() string
	SetPersonID(ID string)
}

type ModelWithTask interface {
	GetTaskName() string
	GetTaskID() string
	SetTaskID(ID string)
}

func (model Activity) GetPersonName() string {
	return model.PersonName
}

func (model Activity) GetPersonID() string {
	return model.PersonID
}

func (model *Activity) SetPersonID(ID string) {
	model.PersonID = ID
}

func (model Activity) GetOrgName() string {
	return model.OrgName
}

func (model Activity) GetOrgID() string {
	return model.OrgID
}

func (model *Activity) SetOrgID(ID string) {
	model.OrgID = ID
}

func (model Activity) GetTaskName() string {
	return model.TaskTitle
}

func (model Activity) GetTaskID() string {
	return model.TaskID
}

func (model *Activity) SetTaskID(ID string) {
	model.TaskID = ID
}

type ActivityField struct {
	Base
	CompanyID          string `json:"company_id"`
	Key                string `json:"key"`
	OrderNr            int    `json:"order_nr"`
	PicklistData       string `json:"picklist_data"`
	FieldType          string `json:"field_type"`
	EditFlag           bool   `json:"edit_flag"`
	IndexVisibleFlag   bool   `json:"index_visible_flag"`
	DetailsVisibleFlag bool   `json:"details_visible_flag"`
	AddVisibleFlag     bool   `json:"add_visible_flag"`
	ImportantFlag      bool   `json:"important_flag"`
	BulkEditAllowed    bool   `json:"bulk_edit_allowed"`
	MandatoryFlag      bool   `json:"mandatory_flag"`
}

type ActivityType struct {
	Base
	CompanyID    string `json:"company_id"`
	KeyString    string `json:"key_string"`
	OrderNr      int    `json:"order_nr"`
	Color        string `json:"color"`
	IsCustomFlag bool   `json:"is_custom_flag"`
}

type Currency struct {
	Base
	CompanyID     string `json:"company_id"`
	Code          string `json:"code"`
	DecimalPoints int    `json:"decimal_points"`
	Symbol        string `json:"symbol"`
	IsCustomFlag  bool   `json:"is_custom_flag"`
}

type ModelWithOrg interface {
	GetOrgName() string
	GetOrgID() string
	SetOrgID(ID string)
}

type Task struct {
	Base
	CreatorUserID          string     `json:"creator_user_id"`
	UserID                 string     `json:"user_id"`
	PersonID               string     `json:"person_id"`
	OrgID                  string     `json:"org_id"`
	StageID                string     `json:"stage_id"`
	Value                  int        `json:"value"`
	Currency               string     `json:"currency"`
	StageChangeTime        *time.Time `json:"stage_change_time"`
	Status                 string     `json:"status"`
	LostReason             string     `json:"lost_reason"`
	VisibleTo              int        `json:"visible_to"`
	CloseTime              *time.Time `json:"close_time"`
	WorkflowID             string     `json:"workflow_id"`
	WonTime                *time.Time `json:"won_time"`
	FirstWonTime           *time.Time `json:"first_won_time"`
	LostTime               *time.Time `json:"lost_time"`
	ExpectedCloseDate      *time.Time `json:"expected_close_date"`
	StageOrderNr           int        `json:"stage_order_nr"`
	FormattedValue         string     `json:"formatted_value"`
	RottenTime             *time.Time `json:"rotten_time"`
	WeightedValue          int        `json:"weighted_value"`
	FormattedWeightedValue string     `json:"formatted_weighted_value"`
	CCEmail                string     `json:"cc_email"`
	OrgHidden              bool       `json:"org_hidden"`
	PersonHidden           bool       `json:"person_hidden"`
	CompanyID              string     `json:"company_id"`

	LastIncomingMailTime *time.Time `json:"last_incoming_mail_time"`
	LastOutgoingMailTime *time.Time `json:"last_outgoing_mail_time"`

	NextActivityDate *time.Time `json:"next_activity_date"`
	NextActivityID   string     `json:"next_activity_id"`
	LastActivityID   string     `json:"last_activity_id"`
	LastActivityDate *time.Time `json:"last_activity_date"`

	NextActivitySubject  string `json:"next_activity_subject"`
	NextActivityType     string `json:"next_activity_type"`
	NextActivityDuration string `json:"next_activity_duration"`
	NextActivityNote     string `json:"next_activity_note"`

	OwnerName                string `json:"owner_name"`
	PersonName               string `json:"person_name"`
	OrgName                  string `json:"org_name"`
	ProductsCount            int    `json:"products_count"`
	FilesCount               int    `json:"files_count"`
	NotesCount               int    `json:"notes_count"`
	FollowersCount           int    `json:"followers_count"`
	EmailMessagesCount       int    `json:"email_messages_count"`
	ActivitiesCount          int    `json:"activities_count"`
	DoneActivitiesCount      int    `json:"done_activities_count"`
	UndoneActivitiesCount    int    `json:"undone_activities_count"`
	ReferenceActivitiesCount int    `json:"reference_activities_count"`
	ParticipantsCount        int    `json:"participants_count"`
	StageName                string `json:"stage_name"`
}

func (model Task) GetPersonName() string {
	return model.PersonName
}

func (model Task) GetPersonID() string {
	return model.PersonID
}

func (model *Task) SetPersonID(ID string) {
	model.PersonID = ID
}

func (model Task) GetOrgName() string {
	return model.OrgName
}

func (model Task) GetOrgID() string {
	return model.OrgID
}

func (model *Task) SetOrgID(ID string) {
	model.OrgID = ID
}

type TaskField struct {
	Base
	CompanyID          string `json:"company_id"`
	Key                string `json:"key"`
	OrderNr            int    `json:"order_nr"`
	PicklistData       string `json:"picklist_data"`
	FieldType          string `json:"field_type"`
	EditFlag           bool   `json:"edit_flag"`
	IndexVisibleFlag   bool   `json:"index_visible_flag"`
	DetailsVisibleFlag bool   `json:"details_visible_flag"`
	AddVisibleFlag     bool   `json:"add_visible_flag"`
	ImportantFlag      bool   `json:"important_flag"`
	BulkEditAllowed    bool   `json:"bulk_edit_allowed"`
	MandatoryFlag      bool   `json:"mandatory_flag"`
}

type File struct {
	Base
	UserID         string `json:"user_id"`
	TaskID         string `json:"task_id"`
	PersonID       string `json:"person_id"`
	OrgID          string `json:"org_id"`
	ProductID      string `json:"product_id"`
	EmailMessageID string `json:"email_message_id"`
	ActivityID     string `json:"activity_id"`
	NoteID         string `json:"note_id"`
	LogID          string `json:"log_id"`
	FileName       string `json:"file_name"`
	FileType       string `json:"file_type"`
	FileSize       int    `json:"file_size"`
	InlineFlag     bool   `json:"inline_flag"`
	RemoteLocation string `json:"remote_location"`
	RemoteID       string `json:"remote_id"`
	CID            string `json:"cid"`
	S3Bucket       string `json:"s3_bucket"`
	MailMessageID  string `json:"mail_message_id"`
	URL            string `json:"url"`
	Description    string `json:"description"`

	TaskName    string `json:"task_name"`
	PersonName  string `json:"person_name"`
	OrgName     string `json:"org_name"`
	ProductName string `json:"product_name"`
}

type Filter struct {
	Base
	CompanyID     string `json:"company_id"`
	Type          string `json:"type"`
	TemporaryFlag bool   `json:"temporary_flag"`
	UserID        string `json:"user_id"`
	VisibleTo     string `json:"visible_to"`
	CustomViewID  string `json:"custom_view_id"`
}

type Goal struct {
	Base
	CompanyID       string     `json:"company_id"`
	UserID          string     `json:"user_id"`
	StageID         string     `json:"stage_id"`
	ActiveGoalID    string     `json:"active_goal_id"`
	Period          string     `json:"period"`
	Expected        int        `json:"expected"`
	GoalType        string     `json:"goal_type"`
	ExpectedSum     int        `json:"expected_sum"`
	Currency        string     `json:"currency"`
	ExpectedType    string     `json:"expected_type"`
	CreatedByUserID string     `json:"created_by_user_id"`
	WorkflowID      string     `json:"workflow_id"`
	MasterExpected  int        `json:"master_expected"`
	Delivered       int        `json:"delivered"`
	DeliveredSum    string     `json:"delivered_sum"`
	PeriodStart     *time.Time `json:"period_start"`
	PeriodEnd       *time.Time `json:"period_end"`

	UserName string `json:"user_name"`
}

type Note struct {
	Base
	CompanyID                string `json:"company_id"`
	UserID                   string `json:"user_id"`
	TaskID                   string `json:"task_id"`
	PersonID                 string `json:"person_id"`
	OrgID                    string `json:"org_id"`
	PinnedToTaskFlag         bool   `json:"pinned_to_task_flag"`
	PinnedToPersonFlag       bool   `json:"pinned_to_person_flag"`
	PinnedToOrganizationFlag bool   `json:"pinned_to_organization_flag"`
	LastUpdateUserID         string `json:"last_update_user_id"`

	OrganizationName string `json:"organization"`
	PersonName       string `json:"person_name"`
	TaskName         string `json:"task_name"`
	UserName         string `json:"user_name"`
}

type NoteField struct {
	Base
	CompanyID     string `json:"company_id"`
	Key           string `json:"key"`
	FieldType     string `json:"field_type"`
	EditFlag      bool   `json:"edit_flag"`
	MandatoryFlag bool   `json:"mandatory_flag"`
}

type Category struct {
	Base
}

type Organization struct {
	Base
	CompanyID              string `json:"company_id"`
	OwnerID                string `json:"owner_id"`
	CategoryID             string `json:"category_id"`
	PictureID              string `json:"picture_id"`
	CountryCode            string `json:"country_code"`
	FirstChar              string `json:"first_char"`
	VisibleTo              string `json:"visible_to"`
	Address                string `json:"address"`
	AddressSubpremise      string `json:"address_subpremise"`
	AddressStreetNumber    string `json:"address_street_number"`
	AddressRoute           string `json:"address_route"`
	AddressSublocality     string `json:"address_sublocality"`
	AddressLocality        string `json:"address_locality"`
	AddressAdminAreaLevel1 string `json:"address_admin_area_level_1"`
	AddressAdminAreaLevel2 string `json:"address_admin_area_level_2"`
	AddressCountry         string `json:"address_country"`
	AddressPostalCode      string `json:"address_postal_code"`
	CCEmail                string `json:"cc_email"`

	NextActivityDate *time.Time `json:"next_activity_date"`
	NextActivityID   string     `json:"next_activity_id"`
	LastActivityID   string     `json:"last_activity_id"`
	LastActivityDate *time.Time `json:"last_activity_date"`

	AddressFormattedAddress  string `json:"address_formatted_address"`
	OpenTasksCount           int    `json:"open_tasks_count"`
	RelatedOpenTasksCount    int    `json:"related_open_tasks_count"`
	ClosedTasksCount         int    `json:"closed_tasks_count"`
	RelatedClosedTasksCount  int    `json:"related_closed_tasks_count"`
	EmailMessagesCount       int    `json:"email_messages_count"`
	PeopleCount              int    `json:"people_count"`
	ActivitiesCount          int    `json:"activities_count"`
	DoneActivitiesCount      int    `json:"done_activities_count"`
	UndoneActivitiesCount    int    `json:"undone_activities_count"`
	ReferenceActivitiesCount int    `json:"reference_activities_count"`
	FilesCount               int    `json:"files_count"`
	NotesCount               int    `json:"notes_count"`
	FollowersCount           int    `json:"followers_count"`
	WonTasksCount            int    `json:"won_tasks_count"`
	RelatedWonTasksCount     int    `json:"related_won_tasks_count"`
	LostTasksCount           int    `json:"lost_tasks_count"`
	RelatedLostTasksCount    int    `json:"related_lost_tasks_count"`
	OwnerName                string `json:"owner_name"`
}

type OrganizationField struct {
	Base
	CompanyID          string `json:"company_id"`
	Key                string `json:"key"`
	OrderNr            int    `json:"order_nr"`
	PicklistData       string `json:"picklist_data"`
	FieldType          string `json:"field_type"`
	EditFlag           bool   `json:"edit_flag"`
	IndexVisibleFlag   bool   `json:"index_visible_flag"`
	DetailsVisibleFlag bool   `json:"details_visible_flag"`
	AddVisibleFlag     bool   `json:"add_visible_flag"`
	ImportantFlag      bool   `json:"important_flag"`
	BulkEditAllowed    bool   `json:"bulk_edit_allowed"`
	UseField           string `json:"use_field"`
	Link               string `json:"link"`
	MandatoryFlag      bool   `json:"mandatory_flag"`
}

type OrganizationRelationship struct {
	Base
	CompanyID              string `json:"company_id"`
	Type                   string `json:"type"`
	RelOwnerOrgID          string `json:"rel_owner_org_id"`
	RelLinkedOrgID         string `json:"rel_linked_org_id"`
	CalculatedType         string `json:"calculated_type"`
	CalculatedRelatedOrgID string `json:"calculated_related_org_id"`

	RelatedOrganizationName string `json:"related_organization_name"`
}

type Contact struct {
	Base
	PersonID string `json:"person_id"`
	Primary  bool   `json:"primary"`
	Type     string `json:"type"`
}

func (c *Contact) DetectType() {
	if strings.Contains(c.Name, "@") {
		c.Type = "email"
	} else {
		c.Type = "phone"
	}
}

type Person struct {
	Base
	CompanyID string `json:"company_id"`
	OwnerID   string `json:"owner_id"`
	OrgID     string `json:"org_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FirstChar string `json:"first_char"`
	VisibleTo string `json:"visible_to"`
	PictureID string `json:"picture_id"`
	CCEmail   string `json:"cc_email"`

	Phone string `json:"phone"`
	Email string `json:"email"`

	NextActivityDate     *time.Time `json:"next_activity_date"`
	NextActivityID       string     `json:"next_activity_id"`
	LastActivityID       string     `json:"last_activity_id"`
	LastActivityDate     *time.Time `json:"last_activity_date"`
	LastIncomingMailTime *time.Time `json:"last_incoming_mail_time"`
	LastOutgoingMailTime *time.Time `json:"last_outgoing_mail_time"`

	OpenTasksCount              int    `json:"open_tasks_count"`
	RelatedOpenTasksCount       int    `json:"related_open_tasks_count"`
	ClosedTasksCount            int    `json:"closed_tasks_count"`
	RelatedClosedTasksCount     int    `json:"related_closed_tasks_count"`
	ParticipantOpenTasksCount   int    `json:"participant_open_tasks_count"`
	ParticipantClosedTasksCount int    `json:"participant_closed_tasks_count"`
	EmailMessagesCount          int    `json:"email_messages_count"`
	ActivitiesCount             int    `json:"activities_count"`
	DoneActivitiesCount         int    `json:"done_activities_count"`
	UndoneActivitiesCount       int    `json:"undone_activities_count"`
	ReferenceActivitiesCount    int    `json:"reference_activities_count"`
	FilesCount                  int    `json:"files_count"`
	NotesCount                  int    `json:"notes_count"`
	FollowersCount              int    `json:"followers_count"`
	WonTasksCount               int    `json:"won_tasks_count"`
	RelatedWonTasksCount        int    `json:"related_won_tasks_count"`
	LostTasksCount              int    `json:"lost_tasks_count"`
	RelatedLostTasksCount       int    `json:"related_lost_tasks_count"`
	OrgName                     string `json:"org_name"`
	OwnerName                   string `json:"owner_name"`
}

func (model Person) GetOrgName() string {
	return model.OrgName
}

func (model Person) GetOrgID() string {
	return model.OrgID
}

func (model *Person) SetOrgID(ID string) {
	model.OrgID = ID
}

type PersonField struct {
	Base
	CompanyID          string `json:"company_id"`
	Key                string `json:"key"`
	OrderNr            int    `json:"order_nr"`
	PicklistData       string `json:"picklist_data"`
	FieldType          string `json:"field_type"`
	EditFlag           bool   `json:"edit_flag"`
	IndexVisibleFlag   bool   `json:"index_visible_flag"`
	DetailsVisibleFlag bool   `json:"details_visible_flag"`
	AddVisibleFlag     bool   `json:"add_visible_flag"`
	ImportantFlag      bool   `json:"important_flag"`
	BulkEditAllowed    bool   `json:"bulk_edit_allowed"`
	UseField           string `json:"use_field"`
	Link               string `json:"link"`
	MandatoryFlag      bool   `json:"mandatory_flag"`
}

type Workflow struct {
	Base
	CompanyID string `json:"company_id"`
	URLTitle  string `json:"url_title"`
	OrderNr   int    `json:"order_nr"`
}

type Price struct {
	Base
	ProductID    string `json:"product_id"`
	Price        int    `json:"price"`
	Currency     string `json:"currency"`
	Cost         int    `json:"cost"`
	OverheadCost string `json:"overhead_cost"`
}

type Product struct {
	Base
	CompanyID  string  `json:"company_id"`
	Code       string  `json:"code"`
	Unit       string  `json:"unit"`
	Tax        int     `json:"tax"`
	Selectable bool    `json:"selectable"`
	FirstChar  string  `json:"first_char"`
	VisibleTo  string  `json:"visible_to"`
	OwnerID    string  `json:"owner_id"`
	Prices     []Price `json:"prices"`

	FilesCount     int `json:"files_count"`
	FollowersCount int `json:"followers_count"`
}

type ProductField struct {
	Base
	CompanyID          string `json:"company_id"`
	Key                string `json:"key"`
	OrderNr            int    `json:"order_nr"`
	PicklistData       string `json:"picklist_data"`
	FieldType          string `json:"field_type"`
	EditFlag           bool   `json:"edit_flag"`
	IndexVisibleFlag   bool   `json:"index_visible_flag"`
	DetailsVisibleFlag bool   `json:"details_visible_flag"`
	AddVisibleFlag     bool   `json:"add_visible_flag"`
	ImportantFlag      bool   `json:"important_flag"`
	BulkEditAllowed    bool   `json:"bulk_edit_allowed"`
	UseField           string `json:"use_field"`
	Link               string `json:"link"`
	MandatoryFlag      bool   `json:"mandatory_flag"`
}

type PushNotification struct {
	Base
	UserID               string `json:"user_id"`
	CompanyID            string `json:"company_id"`
	SubscriptionURL      string `json:"subscription_url"`
	Type                 string `json:"type"`
	Event                string `json:"event"`
	HTTPAuthUser         string `json:"http_auth_user"`
	HTTPAuthPassword     string `json:"http_auth_password"`
	HTTPLastResponseCode string `json:"http_last_response_code"`
}

type Stage struct {
	Base
	OrderNr         int    `json:"order_nr"`
	TaskProbability int    `json:"task_probability"`
	WorkflowID      string `json:"workflow_id"`
	RottenFlag      bool   `json:"rotten_flag"`
	RottenDays      int    `json:"rotten_days"`
}

type Timeline struct {
	Base
	Action                     string `json:"action"`
	UserID                     string `json:"user_id"`
	UnderCompanyID             string `json:"under_company_id"`
	ActivityID                 string `json:"activity_id,omitempty"`
	ActivityFieldID            string `json:"activity_field_id,omitempty"`
	ActivityTypeID             string `json:"activity_type_id,omitempty"`
	CategoryID                 string `json:"category_id,omitempty"`
	CompanyID                  string `json:"company_id,omitempty"`
	CompanyUserID              string `json:"company_user_id,omitempty"`
	ContactID                  string `json:"contact_id,omitempty"`
	CurrencyID                 string `json:"currency_id,omitempty"`
	FileID                     string `json:"file_id,omitempty"`
	FilterID                   string `json:"filter_id,omitempty"`
	GoalID                     string `json:"goal_id,omitempty"`
	NoteFieldID                string `json:"note_field_id,omitempty"`
	NoteID                     string `json:"note_id,omitempty"`
	OrganizationFieldID        string `json:"organization_field_id,omitempty"`
	OrganizationRelationshipID string `json:"organization_relationship_id,omitempty"`
	OrganizationID             string `json:"organization_id,omitempty"`
	PersonID                   string `json:"person_id,omitempty"`
	PersonFieldID              string `json:"person_field_id,omitempty"`
	PriceID                    string `json:"price_id,omitempty"`
	ProductFieldID             string `json:"product_field_id,omitempty"`
	ProductID                  string `json:"product_id,omitempty"`
	PushNotificationID         string `json:"push_notification_id,omitempty"`
	StageID                    string `json:"stage_id,omitempty"`
	TaskFieldID                string `json:"task_field_id,omitempty"`
	TaskID                     string `json:"task_id,omitempty"`
	WorkflowID                 string `json:"workflow_id,omitempty"`
	TimeEntryID                string `json:"time_entry_id,omitempty"`

	UserName string `json:"user_name"`
}

type UserEvent struct {
	Base
	UserID string `json:"user_id"`
}
