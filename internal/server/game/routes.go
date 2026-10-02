package game

import (
	"net/http"

	"github.com/danmrichards/dessego/internal/server/middleware"
)

const routePrefix = "/cgi-bin"

func (s *Server) routes() {
	s.r.HandleFunc("/", middleware.LogRequest(s.l, middleware.DiscardRequestBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Native fallback requires an aborted unknown POST, not an accepted
		// HTTP error response. ErrAbortHandler preserves this without a stack trace.
		if r.Method == http.MethodPost {
			s.l.Warn().Str("path", r.URL.Path).Msg("unsupported native request")
			panic(http.ErrAbortHandler)
		}
		http.NotFound(w, r)
	}))))

	// System routes.
	s.r.HandleFunc(
		routePrefix+"/login.spd",
		middleware.LogRequest(s.l, middleware.DiscardRequestBody(s.loginHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/getTimeMessage.spd",
		middleware.LogRequest(s.l, middleware.DiscardRequestBody(s.timeMsgHandler())),
	)

	// Character/Player routes.
	s.r.HandleFunc(
		routePrefix+"/initializeCharacter.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.initCharacterHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/getQWCData.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.worldTendencyHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/addQWCData.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.addWorldTendencyHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/getMultiPlayGrade.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.characterMPGradeHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/getBloodMessageGrade.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.characterBloodMsgGradeHandler())),
	)

	// Ghost routes.
	s.r.HandleFunc(
		routePrefix+"/getWanderingGhost.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.getGhostHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/setWanderingGhost.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.setGhostHandler())),
	)

	// Blood message routes.
	s.r.HandleFunc(
		routePrefix+"/getBloodMessage.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.getBloodMsgHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/addBloodMessage.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.addBloodMsgHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/deleteBloodMessage.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.deleteBloodMsgHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/updateBloodMessageGrade.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.updateBloodMsgGradeHandler())),
	)

	// Replay routes.
	s.r.HandleFunc(
		routePrefix+"/getReplayList.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.replayListHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/getReplayData.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.getReplayDataHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/addReplayData.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.addReplayDataHandler())),
	)

	// SOS routes.
	s.r.HandleFunc(
		routePrefix+"/getSosData.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.getSosDataHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/addSosData.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.addSosDataHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/checkSosData.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.checkSosDataHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/summonOtherCharacter.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.summonCharacterHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/summonBlackGhost.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.summonBlackGhostHandler())),
	)

	// Multiplayer routes.
	s.r.HandleFunc(
		routePrefix+"/outOfBlock.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.outOfBlockHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/initializeMultiPlay.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.initMultiplayHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/finalizeMultiPlay.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.finaliseMultiplayHandler())),
	)
	s.r.HandleFunc(
		routePrefix+"/updateOtherPlayerGrade.spd",
		middleware.LogRequest(s.l, middleware.LimitRequestBody(s.updateOtherPlayerGradeHandler())),
	)
}
