/*
===========================================================================
ORBIT VM PROTECTOR GPL Source Code
Copyright (C) 2015 Vasileios Anagnostopoulos.
This file is part of the ORBIT VM PROTECTOR Source Code (?ORBIT VM PROTECTOR Source Code?).
ORBIT VM PROTECTOR Source Code is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
ORBIT VM PROTECTOR Source Code is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with ORBIT VM PROTECTOR Source Code.  If not, see <http://www.gnu.org/licenses/>.
In addition, the ORBIT VM PROTECTOR Source Code is also subject to certain additional terms. You should have received a copy of these additional terms immediately following the terms and conditions of the GNU General Public License which accompanied the Doom 3 Source Code.  If not, please request a copy in writing from id Software at the address below.
If you have questions concerning this license or the applicable additional terms, you may contact in writing Vasileios Anagnostopoulos, Campani 3 Street, Athens Greece, POBOX 11252.
===========================================================================
*/
// orbit-vm-protector.go
package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/fithisux/orbit-dc-protector/utilities"
	"github.com/fithisux/orbit-vm-protector/vmprotection"
)

var wa *vmprotection.Watchagent

func orbit_register(request *restful.Request, response *restful.Response) { //stop a stream
	log.Printf("Inside orbit_register")
	params := new(vmprotection.Watchmedata)
	err := request.ReadEntity(params)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	nresp := wa.Register(params)
	response.WriteEntity(nresp)
}

func orbit_update(request *restful.Request, response *restful.Response) { //stop a stream
	log.Printf("Inside orbit_watchme")
	params := new(vmprotection.Watchmedata)
	err := request.ReadEntity(params)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	nresp := wa.Update(params)
	response.WriteEntity(nresp)
}

func orbit_withdraw(request *restful.Request, response *restful.Response) { //stop a stream
	log.Printf("Inside orbit_unwatchme")
	params := new(vmprotection.Watchmedata)
	err := request.ReadEntity(params)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	nresp := wa.Withdraw(params)
	response.WriteEntity(nresp)
}

func orbit_start(request *restful.Request, response *restful.Response) { //stop a stream
	log.Printf("Inside orbit_start")
	nresp := wa.Start()
	response.WriteEntity(nresp)
}

func orbit_stop(request *restful.Request, response *restful.Response) { //stop a stream
	log.Printf("Inside orbit_stop")
	nresp := wa.Stop()
	response.WriteEntity(nresp)
}

func orbit_join(request *restful.Request, response *restful.Response) { //stop a stream
	log.Printf("Inside orbit_join")
	nresp := wa.Join()
	response.WriteEntity(nresp)
}

func orbit_describe(request *restful.Request, response *restful.Response) { //stop a stream
	log.Printf("Inside orbit_describe")
	response.WriteEntity(wa)
}

func heartbeat(reviver chan bool) {
	time.AfterFunc(1*time.Minute, func() {
		reviver <- true
		heartbeat(reviver)
	})
}

func main() {

	conf, err := utilities.Parsetoconf()
	if err != nil {
		panic(err.Error())
	}
	wa = vmprotection.CreateWatchAgent(conf)
	wsContainer := restful.NewContainer()
	log.Printf("Registering")
	ws := new(restful.WebService)
	ws.Path("/ovp").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("").To(orbit_describe))
	ws.Route(ws.POST("/join").To(orbit_join)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/start").To(orbit_start)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/stop").To(orbit_stop)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/register").To(orbit_register)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/update").To(orbit_update)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/withdraw").To(orbit_withdraw)).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	wsContainer.Add(ws)

	// Add container filter to enable CORS
	/*
		cors := restful.CrossOriginResourceSharing{
			ExposeHeaders:  []string{"X-My-Header"},
			AllowedHeaders: []string{"Content-Type"},
			CookiesAllowed: false,
			Container:      wsContainer}
		wsContainer.Filter(cors.Filter)

		// Add container filter to respond to OPTIONS
		wsContainer.Filter(wsContainer.OPTIONSFilter)
	*/

	log.Printf("start listening on localhost:%d\n", conf.Opconfig.Announceport)
	server := &http.Server{Addr: conf.Opconfig.Ovip + ":" + strconv.Itoa(conf.Opconfig.Announceport), Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
