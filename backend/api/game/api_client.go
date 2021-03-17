package game

import (
	"sort"

	"github.com/labstack/echo"
	"github.com/yellia1989/tex-go/tools/log"
	"github.com/yellia1989/tex-web/backend/cfg"
	common "github.com/yellia1989/tex-web/backend/common"
	mid "github.com/yellia1989/tex-web/backend/middleware"
    "strings"
)

type errSimpleInfo struct {
	ErrTime    string `json:"err_time"`
	ErrMessage string `json:"err_info"`
	ErrTimes   uint32 `json:"err_times"`
}

type errSimpleInfoBy []errSimpleInfo

func (a errSimpleInfoBy) Len() int      { return len(a) }
func (a errSimpleInfoBy) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a errSimpleInfoBy) Less(i, j int) bool {
    if a[i].ErrTime > a[j].ErrTime {
        return true
    }

	return a[i].ErrTimes > a[j].ErrTimes
}

type errInfo struct {
	ErrTime    string `json:"err_time"`
	ErrMessage string `json:"err_info"`
	ZoneId     uint32 `json:"zone_id"`
	RoleId     uint32 `json:"role_id"`
}

type errInfoBy []errInfo

func (a errInfoBy) Len() int      { return len(a) }
func (a errInfoBy) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a errInfoBy) Less(i, j int) bool {
	if a[i].ErrTime > a[j].ErrTime {
        return true
    }

    if a[i].ErrMessage > a[j].ErrMessage {
        return true
    }

	return a[i].ZoneId < a[j].ZoneId
}

func ErrInfo(c echo.Context) error {
	ctx := c.(*mid.Context)
	startTime := ctx.QueryParam("startTime")
	endTime := ctx.QueryParam("endTime")

	if startTime == "" || endTime == "" {
		return ctx.SendError(-1, "参数非法")
	}

	db := cfg.LogDb

	sql := "SELECT time, message FROM client_err "
	sql += "WHERE time BETWEEN '" + startTime + "' AND '" + endTime + "'"

	log.Infof("sql: %s", sql)

	rows, err := db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	slSimpleErrInfo := make([]errSimpleInfo, 0)
	for rows.Next() {
		var r errSimpleInfo
		var sErrTime string
		if err := rows.Scan(&sErrTime, &r.ErrMessage); err != nil {
			return err
		}

        r.ErrMessage = strings.Replace(r.ErrMessage, "\n", "<br>", -1)

		// 日期按天排序
		timeFormat := "2006-01-02 15:04:05"
		tErrtime := common.ParseTimeInLocal(timeFormat, sErrTime)

		r.ErrTime = tErrtime.Format("2006-01-02")
		r.ErrTimes += 1

		bFind := false
		for k, v := range slSimpleErrInfo {
			if len(slSimpleErrInfo) != 0 && v.ErrTime == r.ErrTime && v.ErrMessage == r.ErrMessage {
				slSimpleErrInfo[k].ErrTimes += 1
				bFind = true
			}
		}

		if !bFind {
			slSimpleErrInfo = append(slSimpleErrInfo, r)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	sort.Sort(errSimpleInfoBy(slSimpleErrInfo))

	return ctx.SendArray(slSimpleErrInfo, len(slSimpleErrInfo))
}

func ErrDetail(c echo.Context) error {
	ctx := c.(*mid.Context)
	sErrInfo := ctx.QueryParam("ErrInfo")
	sErrTime := ctx.QueryParam("ErrTime")

	if sErrInfo == "" || sErrTime == "" {
		return ctx.SendError(-1, "参数非法")
	}

    // 时间转换
	timeFormat := "2006-01-02 15:04:05"
    timeFormatDay := "2006-01-02"
	tErrtime := common.ParseTimeInLocal(timeFormatDay, sErrTime)
    log.Infof("errTime: %s", tErrtime.Format(timeFormat))
	endTime := tErrtime.AddDate(0, 0, 1)

	sBeginTime := tErrtime.Format(timeFormat)
	sEndTime := endTime.Format(timeFormat)

    sErrInfo = strings.Replace(sErrInfo, "<br>", "\n", -1)

	db := cfg.LogDb
	sql := "SELECT time, message, zoneid, roleid FROM client_err "
	sql += "WHERE time BETWEEN '" + sBeginTime + "' AND '" + sEndTime + "' AND message = '" + sErrInfo + "'"

	log.Infof("sql: %s", sql)

	rows, err := db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	slErrInfo := make([]errInfo, 0)

	for rows.Next() {
		var r errInfo
		if err := rows.Scan(&r.ErrTime, &r.ErrMessage, &r.ZoneId, &r.RoleId); err != nil {
			return err
		}
        r.ErrMessage = strings.Replace(r.ErrMessage, "\n", "<br>", -1)

		slErrInfo = append(slErrInfo, r)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	sort.Sort(errInfoBy(slErrInfo))

	return ctx.SendArray(slErrInfo, len(slErrInfo))
}
