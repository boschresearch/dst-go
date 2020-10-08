// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"context"

	psync "perun.network/go-perun/pkg/sync"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/app/payment"
)

// PayChServer represents a grpc server that implements payment channel API.
type PayChServer struct {
	n perun.NodeAPI

	// The mutex should be used when accessing the map data structures.
	psync.Mutex

	// These maps are used to hold an signal channel for each active subscription.
	// When a subscription is registered, subscribe function will add an entry to the
	// map corresponding to the subscription type.
	// The unsubscribe call should retrieve the channel from the map and close it, which
	// will signal the subscription routine to end.
	//
	// chProposalsNotif works on per session basis and hence this is a map
	// of session id to signaling channel.
	// chUpdatesNotif works on a per channel basis and hence this is a map of session id to
	// channel id to signaling channel.

	chProposalsNotif map[string]chan bool
	chUpdatesNotif   map[string]map[string]chan bool
}

// NewPayChServer returns a new grpc server that can server the payment channel API.
func NewPayChServer(n perun.NodeAPI) *PayChServer {
	return &PayChServer{
		n:                n,
		chProposalsNotif: make(map[string]chan bool),
		chUpdatesNotif:   make(map[string]map[string]chan bool),
	}
}

// GetConfig wraps node.GetConfig.
func (a *PayChServer) GetConfig(context.Context, *pb.GetConfigReq) (*pb.GetConfigResp, error) {
	cfg := a.n.GetConfig()
	return &pb.GetConfigResp{
		ChainAddress:       cfg.ChainURL,
		AdjudicatorAddress: cfg.Adjudicator,
		AssetAddress:       cfg.Asset,
		CommTypes:          cfg.CommTypes,
		ContactTypes:       cfg.ContactTypes,
	}, nil
}

// Time wraps node.Time.
func (a *PayChServer) Time(context.Context, *pb.TimeReq) (*pb.TimeResp, error) {
	return &pb.TimeResp{
		Time: a.n.Time(),
	}, nil
}

// Help wraps node.Help.
func (a *PayChServer) Help(context.Context, *pb.HelpReq) (*pb.HelpResp, error) {
	return &pb.HelpResp{
		Apis: a.n.Help(),
	}, nil
}

// OpenSession wraps node.OpenSession.
func (a *PayChServer) OpenSession(ctx context.Context, req *pb.OpenSessionReq) (*pb.OpenSessionResp, error) {
	errResponse := func(err error) *pb.OpenSessionResp {
		return &pb.OpenSessionResp{
			Response: &pb.OpenSessionResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sessionID, err := a.n.OpenSession(req.ConfigFile)
	if err != nil {
		return errResponse(err), nil
	}

	a.Lock()
	a.chUpdatesNotif[sessionID] = make(map[string]chan bool)
	a.Unlock()

	return &pb.OpenSessionResp{
		Response: &pb.OpenSessionResp_MsgSuccess_{
			MsgSuccess: &pb.OpenSessionResp_MsgSuccess{
				SessionID: sessionID,
			},
		},
	}, nil
}

// AddContact wraps session.AddContact.
func (a *PayChServer) AddContact(ctx context.Context, req *pb.AddContactReq) (*pb.AddContactResp, error) {
	errResponse := func(err error) *pb.AddContactResp {
		return &pb.AddContactResp{
			Response: &pb.AddContactResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	err = sess.AddContact(perun.Peer{
		Alias:              req.Peer.Alias,
		OffChainAddrString: req.Peer.OffChainAddress,
		CommAddr:           req.Peer.CommAddress,
		CommType:           req.Peer.CommType,
	})
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.AddContactResp{
		Response: &pb.AddContactResp_MsgSuccess_{
			MsgSuccess: &pb.AddContactResp_MsgSuccess{
				Success: true,
			},
		},
	}, nil
}

// GetContact wraps session.GetContact.
func (a *PayChServer) GetContact(ctx context.Context, req *pb.GetContactReq) (*pb.GetContactResp, error) {
	errResponse := func(err error) *pb.GetContactResp {
		return &pb.GetContactResp{
			Response: &pb.GetContactResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	peer, err := sess.GetContact(req.Alias)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.GetContactResp{
		Response: &pb.GetContactResp_MsgSuccess_{
			MsgSuccess: &pb.GetContactResp_MsgSuccess{
				Peer: &pb.Peer{
					Alias:           peer.Alias,
					OffChainAddress: peer.OffChainAddrString,
					CommAddress:     peer.CommAddr,
					CommType:        peer.CommType,
				},
			},
		},
	}, nil
}

// OpenPayCh wraps session.OpenPayCh.
func (a *PayChServer) OpenPayCh(ctx context.Context, req *pb.OpenPayChReq) (*pb.OpenPayChResp, error) {
	errResponse := func(err error) *pb.OpenPayChResp {
		return &pb.OpenPayChResp{
			Response: &pb.OpenPayChResp_Error{
				Error: &pb.MsgError{Error: err.Error()},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	openingBalInfo := FromGrpcBalInfo(req.OpeningBalInfo)
	payChInfo, err := payment.OpenPayCh(ctx, sess, openingBalInfo, req.ChallengeDurSecs)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.OpenPayChResp{
		Response: &pb.OpenPayChResp_MsgSuccess_{
			MsgSuccess: &pb.OpenPayChResp_MsgSuccess{
				OpenedPayChInfo: &pb.PayChInfo{
					ChID:    payChInfo.ChID,
					BalInfo: ToGrpcBalInfo(payChInfo.BalInfo),
					Version: payChInfo.Version,
				},
			},
		},
	}, nil
}

// GetPayChs wraps session.GetPayChs.
func (a *PayChServer) GetPayChs(ctx context.Context, req *pb.GetPayChsReq) (*pb.GetPayChsResp, error) {
	errResponse := func(err error) *pb.GetPayChsResp {
		return &pb.GetPayChsResp{
			Response: &pb.GetPayChsResp_Error{
				Error: &pb.MsgError{Error: err.Error()},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	openPayChsInfo := payment.GetPayChsInfo(sess)
	if err != nil {
		return errResponse(err), nil
	}
	openPayChsInfoGrpc := make([]*pb.PayChInfo, len(openPayChsInfo))
	for i := 0; i < len(openPayChsInfoGrpc); i++ {
		openPayChsInfoGrpc[i] = &pb.PayChInfo{
			ChID:    openPayChsInfo[i].ChID,
			BalInfo: ToGrpcBalInfo(openPayChsInfo[i].BalInfo),
			Version: openPayChsInfo[i].Version,
		}
	}

	return &pb.GetPayChsResp{
		Response: &pb.GetPayChsResp_MsgSuccess_{
			MsgSuccess: &pb.GetPayChsResp_MsgSuccess{
				OpenPayChsInfo: openPayChsInfoGrpc,
			},
		},
	}, nil
}

// SubPayChProposals wraps session.SubPayChProposals.
func (a *PayChServer) SubPayChProposals(req *pb.SubPayChProposalsReq,
	srv pb.Payment_API_SubPayChProposalsServer) error {
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		// TODO: (mano) Return a error response and not a protocol error
		return errors.WithMessage(err, "cannot register subscription")
	}

	notifier := func(notif payment.PayChProposalNotif) {
		// nolint: govet	// err does not shadow prev declarations as this runs in a different context.
		err := srv.Send(&pb.SubPayChProposalsResp{Response: &pb.SubPayChProposalsResp_Notify_{
			Notify: &pb.SubPayChProposalsResp_Notify{
				ProposalID:       notif.ProposalID,
				OpeningBalInfo:   ToGrpcBalInfo(notif.OpeningBalInfo),
				ChallengeDurSecs: notif.ChallengeDurSecs,
				Expiry:           notif.Expiry,
			},
		}})
		_ = err
		// if err != nil {
		// TODO: (mano) Handle error while sending.
		// }
	}
	err = payment.SubPayChProposals(sess, notifier)
	if err != nil {
		// TODO: (mano) Return a error response and not a protocol error
		return errors.WithMessage(err, "cannot register subscription")
	}

	signal := make(chan bool)
	a.Lock()
	a.chProposalsNotif[req.SessionID] = signal
	a.Unlock()

	<-signal
	return nil
}

// UnsubPayChProposals wraps session.UnsubPayChProposals.
func (a *PayChServer) UnsubPayChProposals(ctx context.Context, req *pb.UnsubPayChProposalsReq) (
	*pb.UnsubPayChProposalsResp, error) {
	errResponse := func(err error) *pb.UnsubPayChProposalsResp {
		return &pb.UnsubPayChProposalsResp{
			Response: &pb.UnsubPayChProposalsResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	err = payment.UnsubPayChProposals(sess)
	if err != nil {
		return errResponse(err), nil
	}

	a.closeGrpcPayChProposalSub(req.SessionID)

	return &pb.UnsubPayChProposalsResp{
		Response: &pb.UnsubPayChProposalsResp_MsgSuccess_{
			MsgSuccess: &pb.UnsubPayChProposalsResp_MsgSuccess{
				Success: true,
			},
		},
	}, nil
}

func (a *PayChServer) closeGrpcPayChProposalSub(sessionID string) {
	a.Lock()
	signal := a.chProposalsNotif[sessionID]
	delete(a.chProposalsNotif, sessionID)
	a.Unlock()
	close(signal)
}

// RespondPayChProposal wraps session.RespondPayChProposal.
func (a *PayChServer) RespondPayChProposal(ctx context.Context, req *pb.RespondPayChProposalReq) (
	*pb.RespondPayChProposalResp, error) {
	errResponse := func(err error) *pb.RespondPayChProposalResp {
		return &pb.RespondPayChProposalResp{
			Response: &pb.RespondPayChProposalResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	openedPayChInfo, err := payment.RespondPayChProposal(ctx, sess, req.ProposalID, req.Accept)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.RespondPayChProposalResp{
		Response: &pb.RespondPayChProposalResp_MsgSuccess_{
			MsgSuccess: &pb.RespondPayChProposalResp_MsgSuccess{
				OpenedPayChInfo: ToGrpcPayChInfo(openedPayChInfo),
			},
		},
	}, nil
}

// CloseSession wraps session.CloseSession. For now, this is a stub.
func (a *PayChServer) CloseSession(context.Context, *pb.CloseSessionReq) (*pb.CloseSessionResp, error) {
	return nil, nil
}

// SendPayChUpdate wraps ch.SendPayChUpdate.
func (a *PayChServer) SendPayChUpdate(ctx context.Context, req *pb.SendPayChUpdateReq) (
	*pb.SendPayChUpdateResp, error) {
	errResponse := func(err error) *pb.SendPayChUpdateResp {
		return &pb.SendPayChUpdateResp{
			Response: &pb.SendPayChUpdateResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	updatedPayChInfo, err := payment.SendPayChUpdate(ctx, ch, req.Payee, req.Amount)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.SendPayChUpdateResp{
		Response: &pb.SendPayChUpdateResp_MsgSuccess_{
			MsgSuccess: &pb.SendPayChUpdateResp_MsgSuccess{
				UpdatedPayChInfo: ToGrpcPayChInfo(updatedPayChInfo),
			},
		},
	}, nil
}

// SubPayChUpdates wraps ch.SubPayChUpdates.
func (a *PayChServer) SubPayChUpdates(req *pb.SubpayChUpdatesReq, srv pb.Payment_API_SubPayChUpdatesServer) error {
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		// TODO: (mano) Return a error response and not a protocol error.
		return errors.WithMessage(err, "cannot register subscription")
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errors.WithMessage(err, "cannot register subscription")
	}

	notifier := func(notif payment.PayChUpdateNotif) {
		// nolint: govet	// err does not shadow prev declarations as this runs in a different context.
		err := srv.Send(&pb.SubPayChUpdatesResp{Response: &pb.SubPayChUpdatesResp_Notify_{
			Notify: &pb.SubPayChUpdatesResp_Notify{
				UpdateID:          notif.UpdateID,
				ProposedPayChInfo: ToGrpcPayChInfo(notif.ProposedPayChInfo),
				Type:              ToGrpcChUpdateType[notif.Type],
				Expiry:            notif.Expiry,
				Error:             notif.Error,
			},
		}})
		_ = err
		// if err != nil {
		// 	// TODO: (mano) Error handling when sending notification.
		// }

		// Close grpc subscription function (SubPayChUpdates) that will be running in the background.
		if perun.ChUpdateTypeClosed == notif.Type {
			a.closeGrpcPayChUpdateSub(req.SessionID, req.ChID)
		}
	}
	err = payment.SubPayChUpdates(ch, notifier)
	if err != nil {
		// TODO: (mano) Error handling when sending notification.
		return errors.WithMessage(err, "cannot register subscription")
	}

	signal := make(chan bool)
	a.Lock()
	a.chUpdatesNotif[req.SessionID][req.ChID] = signal
	a.Unlock()

	<-signal
	return nil
}

// ToGrpcChUpdateType is a helper var that maps enums from ChUpdateType type defined in perun-node
// to ChUpdateType type defined in grpc package.
var ToGrpcChUpdateType = map[perun.ChUpdateType]pb.SubPayChUpdatesResp_Notify_ChUpdateType{
	perun.ChUpdateTypeOpen:   pb.SubPayChUpdatesResp_Notify_open,
	perun.ChUpdateTypeFinal:  pb.SubPayChUpdatesResp_Notify_final,
	perun.ChUpdateTypeClosed: pb.SubPayChUpdatesResp_Notify_closed,
}

// UnsubPayChUpdates wraps ch.UnsubPayChUpdates.
func (a *PayChServer) UnsubPayChUpdates(ctx context.Context, req *pb.UnsubPayChUpdatesReq) (
	*pb.UnsubPayChUpdatesResp, error) {
	errResponse := func(err error) *pb.UnsubPayChUpdatesResp {
		return &pb.UnsubPayChUpdatesResp{
			Response: &pb.UnsubPayChUpdatesResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	err = payment.UnsubPayChUpdates(ch)
	if err != nil {
		return errResponse(err), nil
	}
	a.closeGrpcPayChUpdateSub(req.SessionID, req.ChID)

	return &pb.UnsubPayChUpdatesResp{
		Response: &pb.UnsubPayChUpdatesResp_MsgSuccess_{
			MsgSuccess: &pb.UnsubPayChUpdatesResp_MsgSuccess{
				Success: true,
			},
		},
	}, nil
}

func (a *PayChServer) closeGrpcPayChUpdateSub(sessionID, chID string) {
	a.Lock()
	signal := a.chUpdatesNotif[sessionID][chID]
	delete(a.chUpdatesNotif[sessionID], chID)
	a.Unlock()
	close(signal)
}

// RespondPayChUpdate wraps ch.RespondPayChUpdate.
func (a *PayChServer) RespondPayChUpdate(ctx context.Context, req *pb.RespondPayChUpdateReq) (
	*pb.RespondPayChUpdateResp, error) {
	errResponse := func(err error) *pb.RespondPayChUpdateResp {
		return &pb.RespondPayChUpdateResp{
			Response: &pb.RespondPayChUpdateResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	updatedPayChInfo, err := payment.RespondPayChUpdate(ctx, ch, req.UpdateID, req.Accept)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.RespondPayChUpdateResp{
		Response: &pb.RespondPayChUpdateResp_MsgSuccess_{
			MsgSuccess: &pb.RespondPayChUpdateResp_MsgSuccess{
				UpdatedPayChInfo: ToGrpcPayChInfo(updatedPayChInfo),
			},
		},
	}, nil
}

// GetPayChInfo wraps ch.GetBalInfo.
func (a *PayChServer) GetPayChInfo(ctx context.Context, req *pb.GetPayChInfoReq) (
	*pb.GetPayChInfoResp, error) {
	errResponse := func(err error) *pb.GetPayChInfoResp {
		return &pb.GetPayChInfoResp{
			Response: &pb.GetPayChInfoResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	payChInfo := payment.GetPayChInfo(ch)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.GetPayChInfoResp{
		Response: &pb.GetPayChInfoResp_MsgSuccess_{
			MsgSuccess: &pb.GetPayChInfoResp_MsgSuccess{
				PayChInfo: ToGrpcPayChInfo(payChInfo),
			},
		},
	}, nil
}

// ClosePayCh wraps ch.ClosePayCh.
func (a *PayChServer) ClosePayCh(ctx context.Context, req *pb.ClosePayChReq) (*pb.ClosePayChResp, error) {
	errResponse := func(err error) *pb.ClosePayChResp {
		return &pb.ClosePayChResp{
			Response: &pb.ClosePayChResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	closedPayChInfo, err := payment.ClosePayCh(ctx, ch)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.ClosePayChResp{
		Response: &pb.ClosePayChResp_MsgSuccess_{
			MsgSuccess: &pb.ClosePayChResp_MsgSuccess{
				ClosedPayChInfo: ToGrpcPayChInfo(closedPayChInfo),
			},
		},
	}, nil
}

// ToGrpcPayChInfo is a helper function to convert PayChInfo struct defined in perun-node
// to PayChInfo struct defined in grpc package. It is exported for use in tests.
func ToGrpcPayChInfo(src payment.PayChInfo) *pb.PayChInfo {
	return &pb.PayChInfo{
		ChID:    src.ChID,
		BalInfo: ToGrpcBalInfo(src.BalInfo),
		Version: src.Version,
	}
}

// FromGrpcBalInfo is a helper function to convert BalInfo struct defined in grpc package
// to BalInfo struct defined in perun-node. It is exported for use in tests.
func FromGrpcBalInfo(src *pb.BalInfo) perun.BalInfo {
	return perun.BalInfo{
		Currency: src.Currency,
		Parts:    src.Parts,
		Bal:      src.Bal,
	}
}

// ToGrpcBalInfo is a helper function to convert BalInfo struct defined in perun-node
// to BalInfo struct defined in grpc package. It is exported for use in tests.
func ToGrpcBalInfo(src perun.BalInfo) *pb.BalInfo {
	return &pb.BalInfo{
		Currency: src.Currency,
		Parts:    src.Parts,
		Bal:      src.Bal,
	}
}
