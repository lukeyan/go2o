/**
 * Copyright 2014 @ z3q.net.
 * name :
 * author : jarryliu
 * date : 2013-12-09 20:14
 * description :
 * history :
 */

package dps

import (
	"errors"
	"fmt"
	"go2o/src/core/domain/interface/member"
	"go2o/src/core/domain/interface/partner"
	"go2o/src/core/domain/interface/valueobject"
	"go2o/src/core/dto"
	"go2o/src/core/infrastructure/format"
	"go2o/src/core/query"
	"time"
)

type memberService struct {
	_memberRep member.IMemberRep
	_query     *query.MemberQuery
}

func NewMemberService(rep member.IMemberRep, q *query.MemberQuery) *memberService {
	return &memberService{
		_memberRep: rep,
		_query:     q,
	}
}

func (this *memberService) GetMember(id int) *member.ValueMember {
	v := this._memberRep.GetMember(id)
	if v != nil {
		nv := v.GetValue()
		return &nv
	}
	return nil
}

func (this *memberService) getMember(partnerId, memberId int) (member.IMember, error) {
	m := this._memberRep.GetMember(memberId)
	if m == nil {
		return m, member.ErrNoSuchMember
	}
	if m.GetRelation().RegisterPartnerId != partnerId {
		return m, partner.ErrPartnerNotMatch
	}
	return m, nil
}

func (this *memberService) GetMemberIdByInvitationCode(code string) int {
	return this._memberRep.GetMemberIdByInvitationCode(code)
}

func (this *memberService) SaveMember(v *member.ValueMember) (int, error) {
	if v.Id > 0 {
		return this.updateMember(v)
	}
	return this.createMember(v)
}

func (this *memberService) updateMember(v *member.ValueMember) (int, error) {
	m := this._memberRep.GetMember(v.Id)
	if m == nil {
		return -1, member.ErrNoSuchMember
	}
	if err := m.SetValue(v); err != nil {
		return m.GetAggregateRootId(), err
	}
	return m.Save()
}

func (this *memberService) createMember(v *member.ValueMember) (int, error) {
	m := this._memberRep.CreateMember(v)
	return m.Save()
}

func (this *memberService) SaveRelation(memberId int, cardId string, invitationId, partnerId int) error {
	m := this._memberRep.GetMember(memberId)
	if m == nil {
		return member.ErrNoSuchMember
	}

	rl := m.GetRelation()
	rl.RefereesId = invitationId
	rl.RegisterPartnerId = partnerId
	rl.CardId = cardId

	return m.SaveRelation(rl)
}

func (this *memberService) GetLevel(memberId int) *valueobject.MemberLevel {
	if m := this._memberRep.GetMember(memberId); m != nil {
		return m.GetLevel()
	}
	return nil
}

func (this *memberService) GetRelation(memberId int) *member.MemberRelation {
	return this._memberRep.GetRelation(memberId)
}

// 锁定/解锁会员
func (this *memberService) LockMember(partnerId, id int) (bool, error) {
	m := this._memberRep.GetMember(id)
	if m == nil {
		return false, member.ErrNoSuchMember
	}

	state := m.GetValue().State
	if state == 1 {
		return false, m.Lock()
	}
	return true, m.Unlock()
}

// 登陆
func (this *memberService) Login(partnerId int, usr, pwd string) (bool, *member.ValueMember, error) {
	val := this._memberRep.GetMemberValueByUsr(usr)
	if val == nil {
		val = this._memberRep.GetMemberValueByPhone(usr)
	}
	if val == nil {
		return false, nil, errors.New("会员不存在")
	}

	if val.Pwd != pwd {
		return false, nil, errors.New("会员用户或密码不正确")
	}

	if val.State == 0 {
		return false, nil, errors.New("会员已停用")
	}

	m := this._memberRep.GetMember(val.Id)
	rl := m.GetRelation()

	if partnerId != -1 && rl.RegisterPartnerId != partnerId {
		return false, nil, errors.New("无法登陆:NOT MATCH PARTNER!")
	}

	unix := time.Now().Unix()
	val.LastLoginTime = unix
	val.UpdateTime = unix

	m.SetValue(val)
	m.Save()

	return true, val, nil
}

func (this *memberService) CheckUsr(usr string, memberId int) error {
	if len(usr) < 6 {
		return member.ErrUserLength
	}
	var id int = this._memberRep.GetMemberIdByUser(usr)
	if id == 0 {
		return nil
	} else if memberId != 0 && id == memberId {
		return nil
	}

	return errors.New("用户名已被使用")
}

func (this *memberService) GetAccount(memberId int) *member.AccountValue {
	m := this._memberRep.CreateMember(&member.ValueMember{Id: memberId})
	//m, _ := this._memberRep.GetMember(memberId)
	//m.AddExp(300)
	return m.GetAccount().GetValue()
}

func (this *memberService) GetBank(memberId int) *member.BankInfo {
	m := this._memberRep.CreateMember(&member.ValueMember{Id: memberId})
	b := m.GetBank()
	return &b
}

func (this *memberService) SaveBankInfo(v *member.BankInfo) error {
	m := this._memberRep.CreateMember(&member.ValueMember{Id: v.MemberId})
	return m.SaveBank(v)
}

// 获取返现记录
func (this *memberService) QueryIncomeLog(memberId, page, size int,
	where, orderBy string) (num int, rows []map[string]interface{}) {
	return this._query.QueryIncomeLog(memberId, page, size, where, orderBy)
}

// 查询分页订单
func (this *memberService) QueryPagerOrder(memberId, page, size int,
	where, orderBy string) (num int, rows []map[string]interface{}) {
	return this._query.QueryPagerOrder(memberId, page, size, where, orderBy)
}

/*********** 收货地址 ***********/
func (this *memberService) GetDeliverAddress(memberId int) []*member.DeliverAddress {
	return this._memberRep.GetDeliverAddress(memberId)
}

//获取配送地址
func (this *memberService) GetDeliverAddressById(memberId,
	deliverId int) *member.DeliverAddress {
	m := this._memberRep.CreateMember(&member.ValueMember{Id: memberId})
	v := m.GetDeliver(deliverId).GetValue()
	return &v
}

//保存配送地址
func (this *memberService) SaveDeliverAddress(memberId int, e *member.DeliverAddress) (int, error) {
	m := this._memberRep.CreateMember(&member.ValueMember{Id: memberId})
	var v member.IDeliver
	if e.Id > 0 {
		v = m.GetDeliver(e.Id)
		v.SetValue(e)
	} else {
		v = m.CreateDeliver(e)
	}
	return v.Save()
}

//删除配送地址
func (this *memberService) DeleteDeliverAddress(memberId int, deliverId int) error {
	m := this._memberRep.CreateMember(&member.ValueMember{Id: memberId})
	return m.DeleteDeliver(deliverId)
}

func (this *memberService) ModifyPassword(memberId int, oldPwd, newPwd string) error {
	m := this._memberRep.GetMember(memberId)
	if m != nil {
		return m.ModifyPassword(newPwd, oldPwd)
	}
	return member.ErrNoSuchMember
}

func (this *memberService) ModifyTradePassword(memberId int, oldPwd, newPwd string) error {
	m := this._memberRep.GetMember(memberId)
	if m != nil {
		return m.ModifyTradePassword(newPwd, oldPwd)
	}
	return member.ErrNoSuchMember
}

//判断会员是否由指定会员邀请推荐的
func (this *memberService) IsInvitation(memberId int, invitationMemberId int) bool {
	m := this._memberRep.CreateMember(&member.ValueMember{Id: memberId})
	return m.Invitation().InvitationBy(invitationMemberId)
}

// 获取我邀请的会员及会员邀请的人数
func (this *memberService) GetMyInvitationMembers(memberId int) ([]*member.ValueMember, map[int]int) {
	iv := this._memberRep.CreateMember(&member.ValueMember{Id: memberId}).Invitation()
	return iv.GetMyInvitationMembers(), iv.GetSubInvitationNum()
}

// 获取会员最后更新时间
func (this *memberService) GetMemberLatestUpdateTime(memberId int) int64 {
	return this._memberRep.GetMemberLatestUpdateTime(memberId)
}

// 获取会员汇总信息
func (this *memberService) GetMemberSummary(memberId int) *dto.MemberSummary {
	var m member.IMember = this._memberRep.GetMember(memberId)
	if m != nil {
		mv := m.GetValue()
		acv := m.GetAccount().GetValue()
		lv := m.GetLevel()
		return &dto.MemberSummary{
			Id:             m.GetAggregateRootId(),
			Usr:            mv.Usr,
			Name:           mv.Name,
			Exp:            mv.Exp,
			Level:          mv.Level,
			LevelName:      lv.Name,
			Integral:       acv.Integral,
			Balance:        acv.Balance,
			PresentBalance: acv.PresentBalance,
			UpdateTime:     mv.UpdateTime,
		}
	}
	return nil
}

// 获取余额变动信息
func (this *memberService) GetBalanceInfoById(memberId, infoId int) *member.BalanceInfoValue {
	m := this._memberRep.GetMember(memberId)
	if m == nil {
		return nil
	}
	return m.GetAccount().GetBalanceInfo(infoId)
}

// 充值
func (this *memberService) Charge(partnerId, memberId, chargeType int, title, tradeNo string, amount float32) error {
	m, err := this.getMember(partnerId, memberId)
	if err != nil {
		return err
	}
	return m.GetAccount().Charge(chargeType, title, tradeNo, amount)
}

// 提现
func (this *memberService) SubmitApplyCash(partnerId, memberId int, outTradePwd string,
	applyType int, applyAmount float32) error {
	m, err := this.getMember(partnerId, memberId)
	if err != nil {
		return err
	}

	tradePwd := m.GetValue().TradePwd
	if len(tradePwd) == 0 {
		return member.ErrNotSetTradePwd
	}

	if tradePwd != outTradePwd {
		return member.ErrIncorrectTradePwd
	}

	acc := m.GetAccount()
	var title string
	switch applyType {
	case member.TypeApplyCashToBank:
		title = "提现到银行卡"
	case member.TypeApplyCashToCharge:
		title = "充值账户"
	case member.TypeApplyCashToServiceProvider:
		title = "充值到第三方账户"
	}
	return acc.RequestApplyCash(applyType, title, applyAmount)
}

// 获取最近的提现
func (this *memberService) GetLatestApplyCash(memberId int) *member.BalanceInfoValue {
	return this._query.GetLatestBalanceInfoByKind(memberId, member.KindBalanceApplyCash)
}

// 获取最近的提现描述
func (this *memberService) GetLatestApplyCashText(memberId int) string {
	var latestInfo string
	latestApplyInfo := this.GetLatestApplyCash(memberId)
	if latestApplyInfo != nil {
		var sText string
		switch latestApplyInfo.State {
		case 0:
			sText = "已提交"
		case 1:
			sText = "已审核"
		case 2:
			sText = "被退回"
		case 3:
			sText = "已完成"
		}
		latestInfo = fmt.Sprintf("您于%s申请提现%s,%s。",
			time.Unix(latestApplyInfo.CreateTime, 0).Format("2006-01-02 15:04:05"),
			format.FormatFloat(latestApplyInfo.Amount),
			sText)
	}
	return latestInfo
}

// 确认提现
func (this *memberService) ConfirmApplyCash(partnerId int, memberId int, infoId int, pass bool, remark string) error {
	m, err := this.getMember(partnerId, memberId)
	if err != nil {
		return err
	}
	return m.GetAccount().ConfirmApplyCash(infoId, pass, remark)
}

// 完成提现
func (this *memberService) FinishApplyCash(partnerId, memberId, id int, tradeNo string) error {
	m, err := this.getMember(partnerId, memberId)
	if err != nil {
		return err
	}
	return m.GetAccount().FinishApplyCash(id, tradeNo)
}
