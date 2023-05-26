package status

import (
	"bytes"
	"fmt"
	"text/template"
)

var templateHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Status page</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.6.2/css/bulma.min.css" />
    <script defer src="https://use.fontawesome.com/releases/v5.0.6/js/all.js"></script>
  </head>
  <body>
    <section class="section">
      <div class="container">
<table class="table">
          <thead>
            <tr>
              <th style="width: 150px">Key</th>
              <th style="width: 100px; text-align: center">Status</th>
              <th style="width: 100px; text-align: center">Replicas</th>
              <th style="width: auto">Info</th>
            </tr>
          </thead>
          <tbody>
            {{ range $row := index .ServiceInfo }}
            <tr>
              <td>{{$row.Key}}</td>
              <td style="text-align: center">
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusHealthy) + ` }}
                  <!-- Status: OK -->
                  <span class="icon has-text-success">
                    <i class="fas fa-check-square"></i>
                  </span>
                {{end}}
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusWarning) + ` }}
                  <!-- Status: WARNING -->
                  <span class="icon has-text-warning">
                    <i class="fas fa-exclamation-triangle"></i>
                  </span>
                {{end}}
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusError) + ` }}
                  <!-- Status: FAIL -->
                  <span class="icon has-text-danger">
                    <i class="fas fa-ban"></i>
                  </span>
                {{end}}
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusUnknown) + ` }}
                  <!-- Status: WARNING -->
                  <span class="icon has-text-info">
                    <i class="fas fa-info-circle"></i>
                  </span>
                {{end}}
              </td>
              <td style="text-align: center">{{$row.InstanceCount}}</td>
              <td>{{$row.Info}}</td>
            </tr>
            {{ end }}
          </tbody>
        </table>

        <table class="table">
          <thead>
            <tr>
              <th style="width: 150px">Key</th>
              <th style="width: 100px; text-align: center">Status</th>
              <th style="width: auto">Info</th>
            </tr>
          </thead>
          <tbody>
            {{ range $row := index .Uptime }}
            <tr>
              <td>{{$row.Key}}</td>
              <td style="text-align: center">
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusHealthy) + ` }}
                  <!-- Status: OK -->
                  <span class="icon has-text-success">
                    <i class="fas fa-check-square"></i>
                  </span>
                {{end}}
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusWarning) + ` }}
                  <!-- Status: WARNING -->
                  <span class="icon has-text-warning">
                    <i class="fas fa-exclamation-triangle"></i>
                  </span>
                {{end}}
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusError) + ` }}
                  <!-- Status: CRITICAL -->
                  <span class="icon has-text-danger">
                    <i class="fas fa-ban"></i>
                  </span>
                {{end}}
                {{if eq $row.Status ` + fmt.Sprintf("%d", ProbeStatusUnknown) + ` }}
                  <!-- Status: WARNING -->
                  <span class="icon has-text-info">
                    <i class="fas fa-info-circle"></i>
                  </span>
                {{end}}
              </td>
              <td>{{$row.Info}}</td>
            </tr>
            {{ end }}
          </tbody>
        </table>
      </div>
    </section>
  </body>
</html>`

func RenderTemplate(data interface{}) ([]byte, error) {
	t, err := template.New("").Parse(templateHTML)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
