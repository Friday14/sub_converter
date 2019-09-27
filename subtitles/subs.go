package subtitles

import (
    "cloud.google.com/go/translate"
    "context"
    "fmt"
    "golang.org/x/text/language"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type Sub struct {
    filename  string
    Timelines []Timeline
}

type Timeline struct {
    Id         int
    TimeString string
    Content    string
}

func CreateSub(file string) (*Sub, error) {
    bytes, err := ioutil.ReadFile(file)
    if err != nil {
        return nil, err
    }

    var sub = parseSubFromFile(bytes)
    sub.SetFilename(filepath.Base(file))
    return &sub, nil
}

func parseSubFromFile(bytes []byte) (s Sub) {
    var fileStr = strings.NewReplacer("\r\n", "\n").Replace(string(bytes))
    var re = regexp.MustCompile(`(?ms)(\d+)\n(\d.*?)\n(.*?)\n{2}`)
    var matches = re.FindAllStringSubmatch(fileStr, -1)
    for _, match := range matches {
        id, _ := strconv.ParseInt(match[1], 10, 32)
        timeline := Timeline{Id: int(id), TimeString: match[2], Content: match[3]}
        s.Timelines = append(s.Timelines, timeline)
    }

    return s
}

func (s *Sub) SetFilename(filename string) {
    s.filename = filename
}

func (s *Sub) Filename() string {
    return s.filename
}

func (s *Sub) SaveToFile() error {
    var content string
    for _, timeline := range s.Timelines {
        content += strconv.FormatInt(int64(timeline.Id), 10) + "\n" + timeline.TimeString + "\n" + timeline.Content + "\n\n"
    }

    f, err := os.OpenFile(s.filename, os.O_RDWR|os.O_CREATE, 0755)
    defer f.Close()
    if err != nil {
        return err
    }

    _, err = f.WriteString(content)
    if err != nil {
        return err
    }

    return err
}

func (s *Sub) Translate(lang language.Tag) (newSub *Sub, err error) {
    newSub = &Sub{}
    var ctx = context.Background()
    client, err := translate.NewClient(ctx)
    if err != nil {
        panic(err)
    }
    chunks := s.compactTimelineByChunks(4000)
    var translated []string
    for i, chunk := range chunks {
        transResp, err := client.Translate(ctx, chunk, lang, &translate.Options{
            Format: translate.Text,
        })
        if err != nil {
            return nil, err
        }
        time.Sleep(time.Second * 3)
        var info = fmt.Sprintf("Progress %s: [%d/%d]", s.Filename(), i+1, len(chunks))
        fmt.Println(info)
        for _, v := range transResp {
            translated = append(translated, v.Text)
        }
    }

    timelines := make([]Timeline, 0)
    for idx, content := range translated {
        timelines = append(timelines, Timeline{Id: s.Timelines[idx].Id, Content: content, TimeString: s.Timelines[idx].TimeString})
    }
    newSub.Timelines = timelines
    var fname = s.Filename()
    newSub.SetFilename(fname[0:len(fname)-len(filepath.Ext(fname))] + "_" + lang.String() + filepath.Ext(fname))
    return newSub, err
}

func (s *Sub) compactTimelineByChunks(sizeChunk int) (stack [][]string) {
    curSize := 0
    curChunk := make([]string, 0)
    for _, v := range s.Timelines {
        curSize += len(v.Content)
        curChunk = append(curChunk, v.Content)
        if curSize >= sizeChunk {
            stack = append(stack, curChunk)
            curChunk = make([]string, 0)
            curSize = 0
        }
    }
    if len(curChunk) > 0 {
        stack = append(stack, curChunk)
    }
    return stack
}
