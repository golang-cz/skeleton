package status

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"strings"
// 	"sync"
// 	"text/template"
//
// 	"github.com/golang-cz/skeleton/pkg/status"
// 	"github.com/golang-cz/skeleton/pkg/ws"
// )
//
// type probe struct {
// 	status.Probe
// 	Key string `json:"key"`
// }
//
// type result struct {
// 	status.Result
// 	Key string `json:"key"`
// }
//
// var smsHTMLPage = `<!DOCTYPE html>
// <html>
//   <head>
//     <meta name="viewport" content="width=device-width, initial-scale=1">
//     <title>Status page</title>
//     <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.6.2/css/bulma.min.css" />
//     <script defer src="https://use.fontawesome.com/releases/v5.0.6/js/all.js"></script>
//   </head>
//   <body>
//     <section class="section">
//       <div class="container">
// <table class="table">
//           <thead>
//             <tr>
//               <th style="width: 150px">Client</th>
//               <th style="width: 200px; text-align: center">Total Price</th>
//             </tr>
//           </thead>
//           <tbody>
// 			{{ range $row := index .ClientRows }}
// 			<tr>
// 				<th>{{ $row.Client }}</th>
// 				<th style="text-align: center;">{{ $row.TotalPrice }}</th>
// 			</tr>
// 			{{ end }}
//           </tbody>
//         </table>
//       </div>
//     </section>
//   </body>
// </html>`
//
// var (
// 	serviceProbes = []probe{
// 		{
// 			Key: "Api",
// 			Probe: &status.HealthProbe{
// 				Subject: "api",
// 			},
// 		},
// 	}
//
// 	uptimeProbes = []probe{
// 		{
// 			Key: "ConvoDb",
// 			Probe: &status.Postgres{
// 				GetDB: func() { return data.DB.Session },
// 			},
// 		},
// 	}
// )
//
// func StatusPage(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	results := run(ctx, append(uptimeProbes, serviceProbes...))
//
// 	if strings.Contains(r.Header.Get("Accept"), "application/json") {
// 		ws.JSON(w, 200, results)
// 		return
// 	}
//
// 	i := len(uptimeProbes) // Helper index to split the slice.
// 	statusPage, err := status.RenderTemplate(struct {
// 		Uptime      []result
// 		ServiceInfo []resul/ }
