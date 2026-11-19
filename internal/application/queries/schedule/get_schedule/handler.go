package get_schedule

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Query struct {
	Year           int
	DistrictCode   string // пусто — все территориальные отделы
	DepartmentCode string // пусто — все районы/отделы
	// Caller — кто смотрит: руководитель — только сотрудников своих теротделов.
	Caller auth.Caller
}

type Part struct {
	Period   string
	Days     int
	Approved bool
}

type Row struct {
	UserID     string
	FullName   string
	Position   string
	Department string
	Parts      []Part
	Planned    int // дней отпуска запланировано на год
	Drafts     bool
	Issues     []string
}

type Option struct {
	Code  string
	Title string
}

// DepartmentOption — район/отдел; DistrictCode пуст у республиканских отделов.
type DepartmentOption struct {
	Code         string
	Title        string
	DistrictCode string
}

type Result struct {
	Year           int
	DistrictCode   string
	DepartmentCode string
	Deadline       string // крайний срок утверждения
	Rows           []Row
	WithIssues     int
	ToApprove      int // сотрудников с неутверждёнными частями
	Districts      []Option
	Departments    []DepartmentOption
}

type Handler struct {
	userRepo       ports.UserRepository
	districtRepo   ports.DistrictRepository
	departmentRepo ports.DepartmentRepository
	leave          *leavecalc.Service
}

func NewHandler(userRepo ports.UserRepository, districtRepo ports.DistrictRepository, departmentRepo ports.DepartmentRepository, leave *leavecalc.Service) *Handler {
	return &Handler{userRepo: userRepo, districtRepo: districtRepo, departmentRepo: departmentRepo, leave: leave}
}

// employees — работающие сотрудники территориального отдела и района/отдела (или все).
func (h *Handler) employees(ctx context.Context, q Query) ([]*user.User, error) {
	all, _, err := h.userRepo.FindUsers(ctx, ports.Query{
		Page: 1, Size: 100000,
		Filters: ports.UserFilter{Status: string(user.StatusActive), DistrictCode: q.DistrictCode, DistrictCodes: q.Caller.Scope(), DepartmentCode: q.DepartmentCode},
	})
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}
	users := make([]*user.User, 0, len(all))
	for _, u := range all {
		if q.Caller.Manages(u) {
			users = append(users, u)
		}
	}
	sort.SliceStable(users, func(i, j int) bool {
		if users[i].Department().Title != users[j].Department().Title {
			return users[i].Department().Title < users[j].Department().Title
		}
		return fullName(users[i]) < fullName(users[j])
	})
	return users, nil
}

// Handle — график отпусков на год с проверкой каждого сотрудника.
func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	users, err := h.employees(ctx, q)
	if err != nil {
		return nil, err
	}

	res := &Result{
		Year:           q.Year,
		DistrictCode:   q.DistrictCode,
		DepartmentCode: q.DepartmentCode,
		Deadline:       vacation.ApprovalDeadline(q.Year).Time().Format("02.01.2006"),
	}
	for _, u := range users {
		snap, err := h.leave.Load(ctx, u, yearEnd(q.Year), "")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fullName(u), err)
		}
		parts := vacation.ScheduleParts(snap.Vacations, q.Year, snap.Holidays)

		row := Row{
			UserID:     u.ID().String(),
			FullName:   fullName(u),
			Position:   u.Position().Title(),
			Department: u.Department().Title,
			Issues:     vacation.ScheduleIssues(parts),
		}
		for _, p := range parts {
			row.Parts = append(row.Parts, Part{
				Period:   p.Start.Time().Format("02.01.2006") + " — " + p.End.Time().Format("02.01.2006"),
				Days:     p.Days,
				Approved: p.Approved,
			})
			row.Planned += p.Days
			row.Drafts = row.Drafts || !p.Approved
		}
		if len(row.Issues) > 0 {
			res.WithIssues++
		}
		if row.Drafts {
			res.ToApprove++
		}
		res.Rows = append(res.Rows, row)
	}

	districts, err := h.districtRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("districts: %w", err)
	}
	for _, d := range districts {
		if q.Caller.ManagesDistrict(d.Code()) {
			res.Districts = append(res.Districts, Option{Code: d.Code(), Title: d.Title()})
		}
	}

	departments, err := h.departmentRepo.ListWithDistrict(ctx)
	if err != nil {
		return nil, fmt.Errorf("departments: %w", err)
	}
	for _, d := range departments {
		res.Departments = append(res.Departments, DepartmentOption{Code: d.Department.Code, Title: d.Department.Title, DistrictCode: d.DistrictCode})
	}
	return res, nil
}

// Report — данные для выгрузки графика в Word: строка на сотрудника и рабочий год,
// за который предоставляются дни отпусков, начинающихся в календарном году.
func (h *Handler) Report(ctx context.Context, q Query) (*ports.ScheduleReport, error) {
	users, err := h.employees(ctx, q)
	if err != nil {
		return nil, err
	}

	report := &ports.ScheduleReport{Year: q.Year}
	seenDept := map[string]bool{}
	for _, u := range users {
		snap, err := h.leave.Load(ctx, u, yearEnd(q.Year), "")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fullName(u), err)
		}

		rows := map[string]*ports.ScheduleReportRow{}
		order := make([]string, 0)
		for _, v := range snap.Vacations {
			if !v.Status().IsActive() || !v.Kind().IsAnnual() || v.StartDate().Year() != q.Year {
				continue
			}
			key := v.WorkYearStart().ISO()
			row, ok := rows[key]
			if !ok {
				row = &ports.ScheduleReportRow{Position: u.Position().Title(), FullName: fullName(u), WorkPeriod: workPeriod(snap, v.WorkYearStart().Time())}
				rows[key] = row
				order = append(order, key)
			}
			days := vacation.ChargeableDays(v.StartDate(), v.EndDate(), snap.Holidays)
			dates := v.StartDate().Time().Format("02.01.2006") + " — " + v.EndDate().Time().Format("02.01.2006")
			switch v.Kind() {
			case vacation.KindBase:
				row.BaseDays += days
				row.BaseDates = append(row.BaseDates, dates)
			case vacation.KindIrregular:
				row.IrregularDays += days
				row.IrregularDates = append(row.IrregularDates, dates)
			case vacation.KindSeniority:
				row.SeniorityDays += days
				row.SeniorityDates = append(row.SeniorityDates, dates)
			}
		}
		if len(order) == 0 {
			continue
		}
		sort.Strings(order)
		for _, key := range order {
			report.Rows = append(report.Rows, *rows[key])
		}
		if d := u.Department().Title; d != "" && !seenDept[d] {
			seenDept[d] = true
			report.Departments = append(report.Departments, d)
		}
	}
	return report, nil
}

// yearEnd — 31 декабря: рабочие годы и праздники нужны до конца календарного года.
func yearEnd(year int) date.Date {
	return date.New(year, time.December, 31)
}

func workPeriod(snap *leavecalc.Snapshot, from time.Time) string {
	for _, y := range snap.Stats {
		if y.From.Equal(from) {
			return y.From.Format("02.01.2006") + "-" + y.To.AddDate(0, 0, -1).Format("02.01.2006")
		}
	}
	return from.Format("02.01.2006") + "-" + from.AddDate(1, 0, -1).Format("02.01.2006")
}

func fullName(u *user.User) string {
	return strings.Join(strings.Fields(u.LastName()+" "+u.FirstName()+" "+u.MiddleName()), " ")
}
