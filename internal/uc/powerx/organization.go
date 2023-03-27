package powerx

import (
	"PowerX/internal/types"
	"PowerX/internal/types/errorx"
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OrganizationUseCase struct {
	db *gorm.DB
}

func NewOrganizationUseCase(db *gorm.DB) *OrganizationUseCase {
	return &OrganizationUseCase{
		db: db,
	}
}

func (e *OrganizationUseCase) Init() {
	var count int64
	if err := e.db.Model(&Department{}).Count(&count).Error; err != nil {
		panic(errors.Wrap(err, "init root dep failed"))
	}
	if count == 0 {
		dep := defaultDepartment()
		if err := e.db.Model(&Department{}).Create(&dep).Error; err != nil {
			panic(errors.Wrap(err, "init root dep failed"))
		}
	}
}

const (
	GenderMale   = "male"
	GenderFeMale = "female"
	GenderUnKnow = "un_know"
)

const (
	EmployeeStatusDisabled = "disabled"
	EmployeeStatusEnabled  = "enabled"
)

type Employee struct {
	types.Model
	Account       string `gorm:"unique"`
	Name          string
	NickName      string
	Desc          string
	Position      string
	JobTitle      string
	DepartmentId  int64
	Department    *Department
	MobilePhone   string
	Gender        string
	Email         string
	ExternalEmail string
	Avatar        string
	Password      string
	Status        string `gorm:"index"`
	IsReserved    bool
	IsActivated   bool
}

const defaultCost = bcrypt.MinCost

// 生成哈希密码
func hashPassword(password string) (hashedPwd string, err error) {
	newPassword, err := bcrypt.GenerateFromPassword([]byte(password), defaultCost)
	if err != nil {
		return "", errors.Wrap(err, "gen pwd failed")
	}
	return string(newPassword), nil
}

// 校验密码
func verifyPassword(hashedPwd string, pwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(pwd))
	return err == nil
}

func (e *Employee) HashPassword() (err error) {
	if e.Password != "" {
		e.Password, err = hashPassword(e.Password)
	}
	return nil
}

type Department struct {
	types.Model
	Name        string
	PId         int64
	PDep        *Department `gorm:"foreignKey:PId"`
	LeaderId    int64
	Leader      *Employee     `gorm:"foreignKey:LeaderId"`
	Ancestors   []*Department `gorm:"many2many:department_ancestors;"`
	Desc        string
	PhoneNumber string
	Email       string
	Remark      string
	IsReserved  bool
}

func defaultDepartment() *Department {
	return &Department{
		Name:       "组织架构",
		PId:        0,
		Desc:       "根节点, 别删除",
		IsReserved: true,
	}
}

func (e *OrganizationUseCase) VerifyPassword(hashedPwd string, pwd string) bool {
	return verifyPassword(hashedPwd, pwd)
}

func (e *OrganizationUseCase) CreateEmployee(ctx context.Context, employee *Employee) (err error) {
	// todo handle conflict
	if err := e.db.WithContext(ctx).Create(&employee).Error; err != nil {
		panic(err)
	}
	return nil
}

func (e *OrganizationUseCase) PatchEmployeeByUserId(ctx context.Context, employee *Employee, employeeId int64) error {
	result := e.db.WithContext(ctx).Model(&Employee{}).Where(employee.ID).Updates(&employee)
	if result.Error != nil {
		panic(result.Error)
	}
	if result.RowsAffected == 0 {
		return errorx.WithCause(errorx.ErrBadRequest, "未找到员工")
	}
	return nil
}

type FindManyEmployeesOption struct {
	Ids             []int64
	Accounts        []string
	Names           []string
	LikeName        string
	Emails          []string
	LikeEmail       string
	DepIds          []int64
	Positions       []string
	PhoneNumbers    []string
	LikePhoneNumber string
	Statuses        []string
	PageIndex       int
	PageSize        int
}

func buildFindManyEmployeesQueryNoPage(query *gorm.DB, opt *FindManyEmployeesOption) *gorm.DB {
	if len(opt.Ids) > 0 {
		query.Where("id in ?", opt.Ids)
	}
	if len(opt.Names) > 0 {
		query.Where("name in ?", opt.Names)
	} else if opt.LikeName != "" {
		query.Where("name like ?", fmt.Sprintf("%s%%", opt.LikeName))
	}
	if len(opt.Emails) > 0 {
		query.Where("email in ?", opt.Emails)
	} else if opt.LikeEmail != "" {
		query.Where("email like ?", fmt.Sprintf("%s%%", opt.LikeEmail))
	}
	if len(opt.PhoneNumbers) > 0 {
		query.Where("mobile_phone in ?")
	} else if opt.LikePhoneNumber != "" {
		query.Where("mobile_phone like ?", fmt.Sprintf("%s%%", opt.LikePhoneNumber))
	}
	if len(opt.Positions) > 0 {
		query.Where("position in ?", opt.Positions)
	}
	if len(opt.Accounts) > 0 {
		query.Where("account in ?", opt.Accounts)
	}
	if len(opt.DepIds) > 0 {
		query.Where("? && department_ids", pq.Int64Array(opt.DepIds))
	}
	if len(opt.Statuses) > 0 {
		query.Where("status in ?", opt.Statuses)
	}
	return query
}

func (e *OrganizationUseCase) FindManyEmployeesPage(ctx context.Context, opt *FindManyEmployeesOption) types.Page[*Employee] {
	var employees []*Employee
	var count int64
	query := e.db.WithContext(ctx).Model(&Employee{})

	if opt.PageIndex != 0 && opt.PageSize != 0 {
		query.Offset((opt.PageIndex - 1) * opt.PageSize).Limit(opt.PageSize)
	}
	query = buildFindManyEmployeesQueryNoPage(query, opt)
	if err := query.Count(&count).Error; err != nil {
		panic(errors.Wrap(err, "find employees failed"))
	}
	if err := query.Find(&employees).Error; err != nil {
		panic(errors.Wrap(err, "find employees failed"))
	}
	return types.Page[*Employee]{
		List:      employees,
		PageIndex: opt.PageIndex,
		PageSize:  opt.PageSize,
		Total:     count,
	}
}

type EmployeeLoginOption struct {
	Account     string
	PhoneNumber string
	Email       string
}

func (e *OrganizationUseCase) FindOneEmployeeByLoginOption(ctx context.Context, option *EmployeeLoginOption) (employee *Employee, err error) {
	if *option == (EmployeeLoginOption{}) {
		panic(errors.New("option empty"))
	}

	var queryEmployee Employee
	if option.Account != "" {
		queryEmployee.Account = option.Account
	}
	if option.Email != "" {
		queryEmployee.Email = option.Email
	}
	if option.PhoneNumber != "" {
		queryEmployee.MobilePhone = option.PhoneNumber
	}

	if err = e.db.WithContext(ctx).Where(queryEmployee).First(employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.WithCause(errorx.ErrBadRequest, "用户不存在, 请检查登录信息")
		}
		panic(err)
	}
	return
}

func (e *OrganizationUseCase) FindOneEmployeeById(ctx context.Context, id int64) (employee *Employee, err error) {
	if err = e.db.WithContext(ctx).Where(id).Preload("Department").First(employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.WithCause(errorx.ErrBadRequest, "用户不存在")
		}
		panic(err)
	}
	return
}

func (e *OrganizationUseCase) UpdateEmployeeById(ctx context.Context, employee *Employee, employeeId int64) {
	whereCase := Employee{
		Model: types.Model{
			ID: employeeId,
		},
		IsReserved: false,
	}
	result := e.db.WithContext(ctx).Where(whereCase, "is_reserved").Updates(employee)
	err := result.Error
	if err != nil {
		panic(errors.Wrap(err, "delete employee failed"))
	}
}

func (e *OrganizationUseCase) DeleteEmployeeById(ctx context.Context, id int64) error {
	result := e.db.WithContext(ctx).Where(Employee{IsReserved: false}, "is_reserved").Delete(&Employee{}, id)
	err := result.Error
	if err != nil {
		panic(errors.Wrap(err, "delete employee failed"))
	}
	if result.RowsAffected == 0 {
		return errorx.WithCause(errorx.ErrBadRequest, "删除失败")
	}
	return nil
}

func (e *OrganizationUseCase) FindAllPositions(ctx context.Context) (positions []string) {
	err := e.db.WithContext(ctx).Model(Employee{}).Pluck("position", &positions)
	if err != nil {
		panic(err)
	}
	return positions
}

func (e *OrganizationUseCase) FindOneDepartment(ctx context.Context, id int64) (department *Department, err error) {
	department = &Department{}
	if err := e.db.WithContext(ctx).Preload("Leader").First(department, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.WithCause(errorx.ErrBadRequest, "部门未找到")
		}
		panic(err)
	}
	return department, nil
}

func (e *OrganizationUseCase) CreateDepartment(ctx context.Context, dep *Department) error {
	if dep.PId == 0 {
		return errorx.WithCause(errorx.ErrBadRequest, "必须指定父部门Id")
	}
	db := e.db.WithContext(ctx)
	// 查询父节点
	var pDep *Department
	if err := db.Preload("Ancestors").Where(dep.PId).First(&pDep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorx.WithCause(errorx.ErrBadRequest, "父部门不存在")
		}
		panic(errors.Wrap(err, "query parent Dep failed"))
	}
	for _, ancestor := range pDep.Ancestors {
		dep.Ancestors = append(dep.Ancestors, ancestor)
	}
	dep.Ancestors = append(dep.Ancestors, pDep)

	if err := db.Create(dep).Error; err != nil {
		panic(errors.Wrap(err, "create dep failed"))
	}
	return nil
}

type FindManyDepartmentsOption struct {
	DepIds []int64
}

func (e *OrganizationUseCase) FindManyDepartmentsPage(ctx context.Context, option types.PageOption[FindManyDepartmentsOption]) *types.Page[*Department] {
	var deps []*Department
	var count int64
	query := e.db.WithContext(ctx).Model(Department{})

	if len(option.Option.DepIds) > 0 {
		query.Where(option.Option.DepIds)
	}

	if err := query.Count(&count).Error; err != nil {
		panic(err)
	}
	if option.PageIndex != 0 && option.PageSize != 0 {
		query.Offset((option.PageIndex - 1) * option.PageSize).Limit(option.PageSize)
	}
	if err := query.Find(&deps).Error; err != nil {
		panic(errors.Wrap(err, "query deps failed"))
	}
	return &types.Page[*Department]{
		List:      deps,
		PageIndex: option.PageIndex,
		PageSize:  option.PageSize,
		Total:     count,
	}
}

func (e *OrganizationUseCase) FindManyDepartmentsByRootId(ctx context.Context, rootId int64) (departments []*Department, err error) {
	if err := e.db.WithContext(ctx).Model(Department{}).Preload("Leader").Preload("Ancestors").
		Joins("Ancestors").Find(&departments); err != nil {
		panic(err)
	}
	if len(departments) == 0 {
		return nil, errorx.WithCause(errorx.ErrBadRequest, "根部门不存在")
	}
	return
}

func (e *OrganizationUseCase) FindAllDepartments(ctx context.Context) (departments []*Department) {
	if err := e.db.WithContext(ctx).Preload("Leader").Find(&departments); err != nil {
		panic(err)
	}
	return
}

func (e *OrganizationUseCase) CountEmployeeInDepartmentByIds(ctx context.Context, depIds []int64) (count int64) {
	if err := e.db.WithContext(ctx).Model(Employee{}).Where("department_id in ?", depIds).Count(&count).Error; err != nil {
		panic(err)
	}
	return count
}

func (e *OrganizationUseCase) DeleteDepartmentById(ctx context.Context, id int64) error {
	result := e.db.WithContext(ctx).Delete(Department{}, id)
	if result.Error != nil {
		panic(result.Error)
	}
	if result.RowsAffected == 0 {
		return errorx.WithCause(errorx.ErrBadRequest, "部门不存在")
	}
	return nil
}