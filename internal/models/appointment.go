package models

import "time"

type AppointmentStatus string

const (
	AppointmentPending   AppointmentStatus = "pending"
	AppointmentConfirmed AppointmentStatus = "confirmed"
	AppointmentCancelled AppointmentStatus = "cancelled"
	AppointmentCompleted AppointmentStatus = "completed"
	AppointmentNoShow    AppointmentStatus = "no_show"
)

// Appointment — запись клиента к мастеру на услугу. EndsAt не приходит
// снаружи: его считает сервис из длительности услуги.
type Appointment struct {
	ID         int               `json:"id"`
	CompanyID  int               `json:"company_id"`
	ClientID   int               `json:"client_id"`
	EmployeeID int               `json:"employee_id"`
	ServiceID  int               `json:"service_id"`
	StartsAt   time.Time         `json:"starts_at"`
	EndsAt     time.Time         `json:"ends_at"`
	Status     AppointmentStatus `json:"status"`
	Comment    string            `json:"comment"`

	// Price — цена услуги на момент записи. Прайс потом меняется, а выручка
	// и история клиента должны остаться такими, какими были.
	Price float64 `json:"price"`
}

func (s AppointmentStatus) Valid() bool {
	switch s {
	case AppointmentPending, AppointmentConfirmed, AppointmentCancelled, AppointmentCompleted, AppointmentNoShow:
		return true
	default:
		return false
	}
}

// statusTransitions — куда запись может перейти. Отменённая и завершённая —
// конечные: возврат отменённой в работу занял бы слот, который к этому времени
// мог уйти другому клиенту, в обход проверки пересечений.
var statusTransitions = map[AppointmentStatus][]AppointmentStatus{
	AppointmentPending:   {AppointmentConfirmed, AppointmentCancelled, AppointmentCompleted, AppointmentNoShow},
	AppointmentConfirmed: {AppointmentCancelled, AppointmentCompleted, AppointmentNoShow},
}

func (s AppointmentStatus) CanBecome(next AppointmentStatus) bool {
	if s == next {
		return true
	}

	for _, allowed := range statusTransitions[s] {
		if allowed == next {
			return true
		}
	}

	return false
}

// Blocks сообщает, занимает ли запись время мастера. Отменённая — не занимает,
// иначе освободившийся слот никто не смог бы забрать.
func (s AppointmentStatus) Blocks() bool {
	return s != AppointmentCancelled
}

// Active — запись ещё впереди и с ней можно работать: переносить, отменять.
// Завершённая и неявка уже случились, отменённая — не случится.
func (s AppointmentStatus) Active() bool {
	return s == AppointmentPending || s == AppointmentConfirmed
}

// Happened — статус, который ставится только после начала визита.
func (s AppointmentStatus) Happened() bool {
	return s == AppointmentCompleted || s == AppointmentNoShow
}
