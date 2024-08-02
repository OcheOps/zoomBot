package api

import (
    "net/http"
 //   "path/filepath"

    "github.com/gin-gonic/gin"
    "github.com/OcheOps/zoomBot/internal/bot"
    "github.com/OcheOps/zoomBot/internal/models"
)

type Server struct {
    bot    *bot.Bot
    router *gin.Engine
}

func NewServer(bot *bot.Bot) *Server {
    s := &Server{
        bot:    bot,
        router: gin.Default(),
    }
    s.routes()
    return s
}

func (s *Server) routes() {
    s.router.LoadHTMLGlob("web/templates/*")
    s.router.Static("/static", "./web/static")

    s.router.GET("/", s.handleIndex)
    s.router.GET("/meetings", s.handleListMeetings)
    s.router.POST("/meetings", s.handleAddMeeting)
    s.router.POST("/join", s.handleJoinMeeting)
}

func (s *Server) Run(addr string) error {
    return s.router.Run(addr)
}

func (s *Server) handleIndex(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title": "Zoom Bot",
    })
}

func (s *Server) handleListMeetings(c *gin.Context) {
    meetings, err := s.bot.ListMeetings()
    if err != nil {
        c.HTML(http.StatusInternalServerError, "error.html", gin.H{
            "error": "Failed to fetch meetings",
        })
        return
    }
    c.HTML(http.StatusOK, "meetings.html", gin.H{
        "title":    "Meetings",
        "meetings": meetings,
    })
}

func (s *Server) handleAddMeeting(c *gin.Context) {
    var meeting models.Meeting
    if err := c.ShouldBind(&meeting); err != nil {
        c.HTML(http.StatusBadRequest, "error.html", gin.H{
            "error": "Invalid meeting data",
        })
        return
    }

    if err := s.bot.AddMeeting(&meeting); err != nil {
        c.HTML(http.StatusInternalServerError, "error.html", gin.H{
            "error": "Failed to add meeting",
        })
        return
    }

    c.Redirect(http.StatusSeeOther, "/meetings")
}

func (s *Server) handleJoinMeeting(c *gin.Context) {
    id := c.PostForm("id")
    go s.bot.JoinMeeting(id) // Run asynchronously
    c.Redirect(http.StatusSeeOther, "/meetings")
}