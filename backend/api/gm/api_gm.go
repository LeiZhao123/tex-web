package gm

import (
    "fmt"
    "strings"
    "bytes"
    "github.com/labstack/echo"
    tex "github.com/yellia1989/tex-go/service"
    mid "github.com/yellia1989/tex-web/backend/middleware"
    "github.com/yellia1989/tex-web/backend/api/gm/rpc"
)

var (
    comm = tex.NewCommunicator("tex.mfwregistry.QueryObj@tcp -h 192.168.0.16 -p 2000 -t 3600000")
)

func GameCmd(c echo.Context) error {
    ctx := c.(*mid.Context)
    zoneid := ctx.FormValue("zoneid")
    scmd := ctx.FormValue("cmd")

    if zoneid == "" || scmd == "" {
        return ctx.SendError(-1, "参数非法")
    }

    gamePrx := new(rpc.GameService)
    comm.StringToProxy("aqua.GameServer.GameServiceObj%aqua.zone."+zoneid, gamePrx)

    gfPrx := new(rpc.GFService)
    comm.StringToProxy("aqua.GFServer.GFServiceObj", gfPrx)

    u := ctx.GetUser()

    buff := bytes.Buffer{}
    cmds := strings.Split(scmd, "\n")
    for _, cmd := range cmds {
        cmd := strings.Trim(strings.ReplaceAll(cmd, "   ", ""), " ")

        result := ""
        var ret int32
        var err error
        if zoneid != "0" {
            ret, err = gamePrx.DoGmCmd(u.UserName, cmd, &result)
        } else {
            ret, err = gfPrx.DoGmCmd(u.UserName, cmd, &result)
        }
        if ret != 0 || err != nil {
            result = fmt.Sprintf("ret:%d, err:%s", ret, err.Error())
        }
        buff.WriteString(result+"\n")
    }

    return ctx.SendResponse(buff.String())
}