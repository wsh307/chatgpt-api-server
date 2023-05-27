package service

import (
	"chatgpt-api-server/modules/chatgpt/model"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gmutex"
)

type SessionPair struct {
	Email       string        `json:"email"`
	Session     string        `json:"session"`
	AccessToken string        `json:"Authorization"`
	Lock        *gmutex.Mutex `json:"lock"`
}

var (
	SessionMap = make(map[string]*SessionPair)
)

type ChatgptUserService struct {
	*cool.Service
}

func NewChatgptUserService() *ChatgptUserService {
	return &ChatgptUserService{
		&cool.Service{
			Model: model.NewChatgptUser(),
		},
	}
}

// GetSessionPair 获取session pair
func (s *ChatgptUserService) GetSessionPair(ctx g.Ctx, userToken string, conversationId string, isPlusModel bool) (sessionPair *SessionPair, code int, err error) {
	record, err := cool.DBM(s.Model).Where("userToken", userToken).Where("expireTime>now()").One()
	if err != nil {
		code = 500
		return
	}
	if record.IsEmpty() {
		code = 401
		err = gerror.New("userToken is not exist or exprieTime is out")
		return
	}
	// 检查用户是否有权限
	if isPlusModel {
		if record["isPlus"].Int() != 1 {
			isPlusModel = false
		}
	}
	sessionRecord, err := NewChatgptSessionService().GetSessionByUserToken(ctx, userToken, conversationId, isPlusModel)
	if err != nil {
		code = 500
		return
	}
	if sessionRecord.IsEmpty() {
		code = 404
		err = gerror.New("session is not exist")
		return
	}
	email := sessionRecord["email"].String()
	sessionPair, ok := SessionMap[email]
	if !ok {
		sessionPair = &SessionPair{
			Email:       email,
			Session:     sessionRecord["officialSession"].String(),
			AccessToken: getAccessTokenFromSession(ctx, sessionRecord["officialSession"].String()),
			Lock:        gmutex.New(),
		}
		if sessionPair.AccessToken == "" {
			code = 500
			err = gerror.New("get accessToken error")
			return
		}
		SessionMap[email] = sessionPair
	}
	return
}

// getaccessTokenFromSession 从session中获取authorization
func getAccessTokenFromSession(ctx g.Ctx, session string) (accessToken string) {
	sessionJson := gjson.New(session)
	// g.Dump(sessionJson)

	accessToken = sessionJson.Get("accessToken").String()
	// g.Log().Debug(ctx, "getAccessTokenFromSession", accessToken)

	return
}