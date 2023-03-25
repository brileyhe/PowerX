package department

import (
	"PowerX/internal/uc/powerx"
	"context"

	"PowerX/internal/svc"
	"PowerX/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDepartmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDepartmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDepartmentLogic {
	return &CreateDepartmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDepartmentLogic) CreateDepartment(req *types.CreateDepartmentRequest) (resp *types.CreateDepartmentReply, err error) {
	dep := powerx.Department{
		Name:     req.DepName,
		LeaderId: req.LeaderId,
		Desc:     req.Desc,
		PId:      req.PId,
	}

	err = l.svcCtx.PowerX.Organization.CreateDepartment(l.ctx, &dep)
	if err != nil {
		return nil, err
	}

	return &types.CreateDepartmentReply{
		Id: dep.ID,
	}, nil
}
