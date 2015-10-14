/*#include <config.h>*/
#include "_cgo_export.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <inttypes.h>
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#define VIR_DEBUG(fmt) printf("%s:%d: " fmt "n", __func__, __LINE__)
#define STREQ(a, b) (strcmp(a, b) == 0)
#ifndef ATTRIBUTE_UNUSED
# define ATTRIBUTE_UNUSED __attribute__((__unused__))
#endif
int run = 1;
/* Prototypes */
const char *eventToString(int event);
int myEventAddHandleFunc  (int fd, int event,
virEventHandleCallback cb,
void *opaque,
virFreeCallback ff);
void myEventUpdateHandleFunc(int watch, int event);
int  myEventRemoveHandleFunc(int watch);
int myEventAddTimeoutFunc(int timeout,
virEventTimeoutCallback cb,
void *opaque,
virFreeCallback ff);
void myEventUpdateTimeoutFunc(int timer, int timout);
int myEventRemoveTimeoutFunc(int timer);
int myEventHandleTypeToPollEvent(virEventHandleType events);
virEventHandleType myPollEventToEventHandleType(int events);
void usage(const char *pname);
#ifdef __MINGW32__
char *strdup(const char *str)
{
    size_t len;
    char *newstr;
    if(!str)
    return (char *)NULL;
    len = strlen(str);
    if(len >= ((size_t)-1) / sizeof(char))
    return (char *)NULL;
    newstr = malloc((len+1)*sizeof(char));
    if(!newstr)
    return (char *)NULL;
    memcpy(newstr,str,(len+1)*sizeof(char));
    return newstr;
}
#endif
/* Callback functions */
static void connectClose(virConnectPtr conn ATTRIBUTE_UNUSED,
int reason,
void *opaque ATTRIBUTE_UNUSED)
{
    switch (reason) {
        case VIR_CONNECT_CLOSE_REASON_ERROR:
        fprintf(stderr, "Connection closed due to I/O errorn");
        break;
        case VIR_CONNECT_CLOSE_REASON_EOF:
        fprintf(stderr, "Connection closed due to end of filen");
        break;
        case VIR_CONNECT_CLOSE_REASON_KEEPALIVE:
        fprintf(stderr, "Connection closed due to keepalive timeoutn");
        break;
        case VIR_CONNECT_CLOSE_REASON_CLIENT:
        fprintf(stderr, "Connection closed due to client requestn");
        break;
        default:
        fprintf(stderr, "Connection closed due to unknown reasonn");
        break;
    };
    run = 0;
}
const char *eventToString(int event) {
    const char *ret = "";
    switch ((virDomainEventType) event) {
        case VIR_DOMAIN_EVENT_DEFINED:
        ret = "Defined";
        break;
        case VIR_DOMAIN_EVENT_UNDEFINED:
        ret = "Undefined";
        break;
        case VIR_DOMAIN_EVENT_STARTED:
        ret = "Started";
        break;
        case VIR_DOMAIN_EVENT_SUSPENDED:
        ret = "Suspended";
        break;
        case VIR_DOMAIN_EVENT_RESUMED:
        ret = "Resumed";
        break;
        case VIR_DOMAIN_EVENT_STOPPED:
        ret = "Stopped";
        break;
        case VIR_DOMAIN_EVENT_SHUTDOWN:
        ret = "Shutdown";
        break;
        case VIR_DOMAIN_EVENT_PMSUSPENDED:
        ret = "PMSuspended";
        break;
        case VIR_DOMAIN_EVENT_CRASHED:
        ret = "Crashed";
        break;
    }
    return ret;
}
static const char *eventDetailToString(int event, int detail) {
    const char *ret = "";
    switch ((virDomainEventType) event) {
        case VIR_DOMAIN_EVENT_DEFINED:
        if (detail == VIR_DOMAIN_EVENT_DEFINED_ADDED)
        ret = "Added";
        else if (detail == VIR_DOMAIN_EVENT_DEFINED_UPDATED)
        ret = "Updated";
        break;
        case VIR_DOMAIN_EVENT_UNDEFINED:
        if (detail == VIR_DOMAIN_EVENT_UNDEFINED_REMOVED)
        ret = "Removed";
        break;
        case VIR_DOMAIN_EVENT_STARTED:
        switch ((virDomainEventStartedDetailType) detail) {
            case VIR_DOMAIN_EVENT_STARTED_BOOTED:
            ret = "Booted";
            break;
            case VIR_DOMAIN_EVENT_STARTED_MIGRATED:
            ret = "Migrated";
            break;
            case VIR_DOMAIN_EVENT_STARTED_RESTORED:
            ret = "Restored";
            break;
            case VIR_DOMAIN_EVENT_STARTED_FROM_SNAPSHOT:
            ret = "Snapshot";
            break;
            case VIR_DOMAIN_EVENT_STARTED_WAKEUP:
            ret = "Event wakeup";
            break;
        }
        break;
        case VIR_DOMAIN_EVENT_SUSPENDED:
        switch ((virDomainEventSuspendedDetailType) detail) {
            case VIR_DOMAIN_EVENT_SUSPENDED_PAUSED:
            ret = "Paused";
            break;
            case VIR_DOMAIN_EVENT_SUSPENDED_MIGRATED:
            ret = "Migrated";
            break;
            case VIR_DOMAIN_EVENT_SUSPENDED_IOERROR:
            ret = "I/O Error";
            break;
            case VIR_DOMAIN_EVENT_SUSPENDED_WATCHDOG:
            ret = "Watchdog";
            break;
            case VIR_DOMAIN_EVENT_SUSPENDED_RESTORED:
            ret = "Restored";
            break;
            case VIR_DOMAIN_EVENT_SUSPENDED_FROM_SNAPSHOT:
            ret = "Snapshot";
            break;
            case VIR_DOMAIN_EVENT_SUSPENDED_API_ERROR:
            ret = "API error";
            break;
        }
        break;
        case VIR_DOMAIN_EVENT_RESUMED:
        switch ((virDomainEventResumedDetailType) detail) {
            case VIR_DOMAIN_EVENT_RESUMED_UNPAUSED:
            ret = "Unpaused";
            break;
            case VIR_DOMAIN_EVENT_RESUMED_MIGRATED:
            ret = "Migrated";
            break;
            case VIR_DOMAIN_EVENT_RESUMED_FROM_SNAPSHOT:
            ret = "Snapshot";
            break;
        }
        break;
        case VIR_DOMAIN_EVENT_STOPPED:
        switch ((virDomainEventStoppedDetailType) detail) {
            case VIR_DOMAIN_EVENT_STOPPED_SHUTDOWN:
            ret = "Shutdown";
            break;
            case VIR_DOMAIN_EVENT_STOPPED_DESTROYED:
            ret = "Destroyed";
            break;
            case VIR_DOMAIN_EVENT_STOPPED_CRASHED:
            ret = "Crashed";
            break;
            case VIR_DOMAIN_EVENT_STOPPED_MIGRATED:
            ret = "Migrated";
            break;
            case VIR_DOMAIN_EVENT_STOPPED_SAVED:
            ret = "Saved";
            break;
            case VIR_DOMAIN_EVENT_STOPPED_FAILED:
            ret = "Failed";
            break;
            case VIR_DOMAIN_EVENT_STOPPED_FROM_SNAPSHOT:
            ret = "Snapshot";
            break;
        }
        break;
        case VIR_DOMAIN_EVENT_SHUTDOWN:
        switch ((virDomainEventShutdownDetailType) detail) {
            case VIR_DOMAIN_EVENT_SHUTDOWN_FINISHED:
            ret = "Finished";
            break;
        }
        break;
        case VIR_DOMAIN_EVENT_PMSUSPENDED:
        switch ((virDomainEventPMSuspendedDetailType) detail) {
            case VIR_DOMAIN_EVENT_PMSUSPENDED_MEMORY:
            ret = "Memory";
            break;
            case VIR_DOMAIN_EVENT_PMSUSPENDED_DISK:
            ret = "Disk";
            break;
        }
        break;
        case VIR_DOMAIN_EVENT_CRASHED:
        switch ((virDomainEventCrashedDetailType) detail) {
            case VIR_DOMAIN_EVENT_CRASHED_PANICKED:
            ret = "Panicked";
            break;
        }
        break;
    }
    return ret;
}
static int myDomainEventCallback2(virConnectPtr conn ATTRIBUTE_UNUSED,
virDomainPtr dom,
int event,
int detail,
void *opaque ATTRIBUTE_UNUSED)
{
    printf("%s EVENT: Domain %s(%d) %s %sn", __func__, virDomainGetName(dom),
    virDomainGetID(dom), eventToString(event),
    eventDetailToString(event, detail));
    char * buf = (char *) malloc(VIR_UUID_STRING_BUFLEN);
    int ret= virDomainGetUUIDString(dom,buf);
    if (ret < 0) {
        VIR_DEBUG("Failed to get uuid");
        exit(-1);
    }
    myReportEvent(buf,virDomainGetID(dom),event,detail);
    free(buf);
    return 0;
}
static int myDomainEventRebootCallback(virConnectPtr conn ATTRIBUTE_UNUSED,
virDomainPtr dom,
void *opaque ATTRIBUTE_UNUSED)
{
    printf("%s EVENT: Domain %s(%d) rebootedn", __func__, virDomainGetName(dom),
    virDomainGetID(dom));
    return 0;
}
static void myFreeFunc(void *opaque)
{
    char *str = opaque;
    printf("%s: Freeing [%s]n", __func__, str);
    free(str);
}
/* main test functions */
void usage(const char *pname)
{
    printf("%s urin", pname);
}
static void stop(int sig)
{
    printf("Exiting on signal %dn", sig);
    run = 0;
}
int main1()
{
    int callback2ret = -1;
    int callback3ret = -1;
    if (virInitialize() < 0) {
        fprintf(stderr, "Failed to initialize libvirt");
        return -1;
    }
    if (virEventRegisterDefaultImpl() < 0) {
        virErrorPtr err = virGetLastError();
        fprintf(stderr, "Failed to register event implementation: %sn",
        err && err->message ? err->message: "Unknown error");
        return -1;
    }
    virConnectPtr dconn = NULL;
    dconn = virConnectOpenAuth( NULL,
    virConnectAuthPtrDefault,
    VIR_CONNECT_RO);
    if (!dconn) {
        printf("error openingn");
        return -1;
    }
    virConnectRegisterCloseCallback(dconn,
    connectClose, NULL, NULL);
    VIR_DEBUG("Registering listing domains");
    virDomainPtr *domains;
    size_t i;
    int ret;
    unsigned int flags = VIR_CONNECT_LIST_DOMAINS_RUNNING |
    VIR_CONNECT_LIST_DOMAINS_PERSISTENT;
    ret = virConnectListAllDomains(dconn, &domains, flags);
    if (ret < 0)
    // error();
    exit(-1);
    char * buf = (char *) malloc(VIR_UUID_STRING_BUFLEN);
    for (i = 0; i < ret; i++) {
        ret=virDomainGetUUIDString(domains[i],buf);
        if (ret < 0) {
            VIR_DEBUG("Failed to get uuid");
            exit(-1);
        }
        myReportID(buf,virDomainGetID(domains[i]));
        virDomainFree(domains[i]);
    }
    free(domains);
    free(buf);
    myReportStraight();
    VIR_DEBUG("Registering event cbs");
    /* Add 2 callbacks to prove this works with more than just one */
    callback2ret = virConnectDomainEventRegisterAny(dconn,
    NULL,
    VIR_DOMAIN_EVENT_ID_LIFECYCLE,
    VIR_DOMAIN_EVENT_CALLBACK(myDomainEventCallback2),
    strdup("callback 2"), myFreeFunc);
    callback3ret = virConnectDomainEventRegisterAny(dconn,
    NULL,
    VIR_DOMAIN_EVENT_ID_REBOOT,
    VIR_DOMAIN_EVENT_CALLBACK(myDomainEventRebootCallback),
    strdup("callback reboot"), myFreeFunc);
    if ((callback2ret != -1) &&
    (callback3ret != -1)) {
        if (virConnectSetKeepAlive(dconn, 5, 3) < 0) {
            virErrorPtr err = virGetLastError();
            fprintf(stderr, "Failed to start keepalive protocol: %sn",
            err && err->message ? err->message : "Unknown error");
            run = 0;
        }
        while (run) {
            if (virEventRunDefaultImpl() < 0) {
                virErrorPtr err = virGetLastError();
                fprintf(stderr, "Failed to run event loop: %sn",
                err && err->message ? err->message : "Unknown error");
            }
        }
        VIR_DEBUG("Deregistering event handlers");
        virConnectDomainEventDeregisterAny(dconn, callback2ret);
        virConnectDomainEventDeregisterAny(dconn, callback3ret);
    }
    virConnectUnregisterCloseCallback(dconn, connectClose);
    VIR_DEBUG("Closing connection");
    if (dconn && virConnectClose(dconn) < 0)
    printf("error closingn");
    printf("donen");
    return 0;
}