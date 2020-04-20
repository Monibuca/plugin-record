(function webpackUniversalModuleDefinition(root, factory) {
	if(typeof exports === 'object' && typeof module === 'object')
		module.exports = factory();
	else if(typeof define === 'function' && define.amd)
		define([], factory);
	else if(typeof exports === 'object')
		exports["plugin-record"] = factory();
	else
		root["plugin-record"] = factory();
})((typeof self !== 'undefined' ? self : this), function() {
return /******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};
/******/
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/
/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId]) {
/******/ 			return installedModules[moduleId].exports;
/******/ 		}
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			i: moduleId,
/******/ 			l: false,
/******/ 			exports: {}
/******/ 		};
/******/
/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);
/******/
/******/ 		// Flag the module as loaded
/******/ 		module.l = true;
/******/
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/
/******/
/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;
/******/
/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;
/******/
/******/ 	// define getter function for harmony exports
/******/ 	__webpack_require__.d = function(exports, name, getter) {
/******/ 		if(!__webpack_require__.o(exports, name)) {
/******/ 			Object.defineProperty(exports, name, { enumerable: true, get: getter });
/******/ 		}
/******/ 	};
/******/
/******/ 	// define __esModule on exports
/******/ 	__webpack_require__.r = function(exports) {
/******/ 		if(typeof Symbol !== 'undefined' && Symbol.toStringTag) {
/******/ 			Object.defineProperty(exports, Symbol.toStringTag, { value: 'Module' });
/******/ 		}
/******/ 		Object.defineProperty(exports, '__esModule', { value: true });
/******/ 	};
/******/
/******/ 	// create a fake namespace object
/******/ 	// mode & 1: value is a module id, require it
/******/ 	// mode & 2: merge all properties of value into the ns
/******/ 	// mode & 4: return value when already ns object
/******/ 	// mode & 8|1: behave like require
/******/ 	__webpack_require__.t = function(value, mode) {
/******/ 		if(mode & 1) value = __webpack_require__(value);
/******/ 		if(mode & 8) return value;
/******/ 		if((mode & 4) && typeof value === 'object' && value && value.__esModule) return value;
/******/ 		var ns = Object.create(null);
/******/ 		__webpack_require__.r(ns);
/******/ 		Object.defineProperty(ns, 'default', { enumerable: true, value: value });
/******/ 		if(mode & 2 && typeof value != 'string') for(var key in value) __webpack_require__.d(ns, key, function(key) { return value[key]; }.bind(null, key));
/******/ 		return ns;
/******/ 	};
/******/
/******/ 	// getDefaultExport function for compatibility with non-harmony modules
/******/ 	__webpack_require__.n = function(module) {
/******/ 		var getter = module && module.__esModule ?
/******/ 			function getDefault() { return module['default']; } :
/******/ 			function getModuleExports() { return module; };
/******/ 		__webpack_require__.d(getter, 'a', getter);
/******/ 		return getter;
/******/ 	};
/******/
/******/ 	// Object.prototype.hasOwnProperty.call
/******/ 	__webpack_require__.o = function(object, property) { return Object.prototype.hasOwnProperty.call(object, property); };
/******/
/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";
/******/
/******/
/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(__webpack_require__.s = "fb15");
/******/ })
/************************************************************************/
/******/ ({

/***/ "0e05":
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
/* harmony import */ var _node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_Records_vue_vue_type_style_index_0_id_7c800264_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__("a189");
/* harmony import */ var _node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_Records_vue_vue_type_style_index_0_id_7c800264_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0___default = /*#__PURE__*/__webpack_require__.n(_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_Records_vue_vue_type_style_index_0_id_7c800264_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0__);
/* unused harmony reexport * */
 /* unused harmony default export */ var _unused_webpack_default_export = (_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_Records_vue_vue_type_style_index_0_id_7c800264_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0___default.a); 

/***/ }),

/***/ "39a8":
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
/* harmony import */ var _node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_id_3ce37488_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__("b3f8");
/* harmony import */ var _node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_id_3ce37488_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0___default = /*#__PURE__*/__webpack_require__.n(_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_id_3ce37488_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0__);
/* unused harmony reexport * */
 /* unused harmony default export */ var _unused_webpack_default_export = (_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_id_3ce37488_scoped_true_lang_css___WEBPACK_IMPORTED_MODULE_0___default.a); 

/***/ }),

/***/ "a189":
/***/ (function(module, exports, __webpack_require__) {

// extracted by mini-css-extract-plugin

/***/ }),

/***/ "b3f8":
/***/ (function(module, exports, __webpack_require__) {

// extracted by mini-css-extract-plugin

/***/ }),

/***/ "f6fd":
/***/ (function(module, exports) {

// document.currentScript polyfill by Adam Miller

// MIT license

(function(document){
  var currentScript = "currentScript",
      scripts = document.getElementsByTagName('script'); // Live NodeList collection

  // If browser needs currentScript polyfill, add get currentScript() to the document object
  if (!(currentScript in document)) {
    Object.defineProperty(document, currentScript, {
      get: function(){

        // IE 6-10 supports script readyState
        // IE 10+ support stack trace
        try { throw new Error(); }
        catch (err) {

          // Find the second match for the "at" string to get file src url from stack.
          // Specifically works with the format of stack traces in IE.
          var i, res = ((/.*at [^\(]*\((.*):.+:.+\)$/ig).exec(err.stack) || [false])[1];

          // For all scripts on the page, if src matches or if ready state is interactive, return the script tag
          for(i in scripts){
            if(scripts[i].src == res || scripts[i].readyState == "interactive"){
              return scripts[i];
            }
          }

          // If no match, return null
          return null;
        }
      }
    });
  }
})(document);


/***/ }),

/***/ "fb15":
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
// ESM COMPAT FLAG
__webpack_require__.r(__webpack_exports__);

// CONCATENATED MODULE: ./node_modules/@vue/cli-service/lib/commands/build/setPublicPath.js
// This file is imported into lib/wc client bundles.

if (typeof window !== 'undefined') {
  if (true) {
    __webpack_require__("f6fd")
  }

  var i
  if ((i = window.document.currentScript) && (i = i.src.match(/(.+\/)[^/]+\.js(\?.*)?$/))) {
    __webpack_require__.p = i[1] // eslint-disable-line
  }
}

// Indicate to webpack that this file can be concatenated
/* harmony default export */ var setPublicPath = (null);

// CONCATENATED MODULE: ./node_modules/cache-loader/dist/cjs.js?{"cacheDirectory":"node_modules/.cache/vue-loader","cacheIdentifier":"5b04a8fd-vue-loader-template"}!./node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/App.vue?vue&type=template&id=3ce37488&scoped=true&
var render = function () {var _vm=this;var _h=_vm.$createElement;var _c=_vm._self._c||_h;return _c('div',[_c('mu-tabs',{attrs:{"value":_vm.active1,"indicator-color":"#80deea","inverse":"","center":""},on:{"update:value":function($event){_vm.active1=$event}}},[_c('mu-tab',[_vm._v("直播流")]),_c('mu-tab',[_vm._v("录制的视频")])],1),(_vm.active1==0)?_c('mu-data-table',{attrs:{"columns":_vm.columns,"data":_vm.$store.state.Room,"min-col-width":50},on:{"row-click":function (i,r){ return _vm.isRecording(r)?_vm.stopRecord(r):_vm.record(r); }},scopedSlots:_vm._u([{key:"default",fn:function(scope){return [(_vm.isRecording(scope.row))?_c('td',{staticClass:"is-center"}):_c('td',{staticClass:"is-center"}),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(scope.row.StreamPath))]),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(scope.row.Type||"await"))]),_c('td',{staticClass:"is-center"},[_c('StartTime',{attrs:{"value":scope.row.StartTime}})],1),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(_vm.SoundFormat(scope.row.AudioInfo.SoundFormat)))]),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(_vm.SoundRate(scope.row.AudioInfo.SoundRate)))]),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(scope.row.AudioInfo.SoundType))]),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(_vm.CodecID(scope.row.VideoInfo.CodecID)))]),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(scope.row.VideoInfo.SPSInfo.Width)+"x"+_vm._s(scope.row.VideoInfo.SPSInfo.Height))]),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(scope.row.AudioInfo.PacketCount)+"/"+_vm._s(scope.row.VideoInfo.PacketCount))]),_c('td',{staticClass:"is-center"},[_vm._v(_vm._s(_vm.getSubscriberCount(scope.row)))])]}}],null,false,1124011921)}):_vm._e(),(_vm.active1==1)?_c('Records',{ref:"recordsPanel"}):_vm._e()],1)}
var staticRenderFns = []


// CONCATENATED MODULE: ./src/App.vue?vue&type=template&id=3ce37488&scoped=true&

// CONCATENATED MODULE: ./node_modules/cache-loader/dist/cjs.js?{"cacheDirectory":"node_modules/.cache/vue-loader","cacheIdentifier":"5b04a8fd-vue-loader-template"}!./node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/components/Records.vue?vue&type=template&id=7c800264&scoped=true&
var Recordsvue_type_template_id_7c800264_scoped_true_render = function () {var _vm=this;var _h=_vm.$createElement;var _c=_vm._self._c||_h;return _c('div',{staticClass:"records"},_vm._l((_vm.data),function(item){return _c('mu-card',{key:item},[_c('mu-card-title',{attrs:{"title":item.Path,"sub-title":_vm.unitFormat(item.Size)+' '+_vm.toDurationStr(item.Duration)}}),_c('mu-card-actions',[_c('mu-button',{attrs:{"icon":"","small":""},on:{"click":function($event){return _vm.play(item)}}},[_c('mu-icon',{attrs:{"value":"play_arrow"}})],1),_c('mu-button',{attrs:{"icon":"","small":""},on:{"click":function($event){return _vm.deleteFlv(item)}}},[_c('mu-icon',{attrs:{"value":"delete_forever"}})],1)],1)],1)}),1)}
var Recordsvue_type_template_id_7c800264_scoped_true_staticRenderFns = []


// CONCATENATED MODULE: ./src/components/Records.vue?vue&type=template&id=7c800264&scoped=true&

// CONCATENATED MODULE: ./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/components/Records.vue?vue&type=script&lang=js&
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//

/* harmony default export */ var Recordsvue_type_script_lang_js_ = ({
    data() {
        return {
            data: []
        };
    },
    methods: {
        play(item) {
            this.ajax.get(
                "/record/flv/play",
                { streamPath: item.Path.replace(".flv", "") },
                x => {
                    if (x == "success") {
                        this.onVisible(true);
                        this.$toast.success("开始发布");
                    } else {
                        this.$toast.error(x);
                    }
                }
            );
        },
        deleteFlv(item) {
            this.$confirm("是否删除Flv文件", "提示").then(result => {
                if (result) {
                    return this.ajax.get(
                        "/record/flv/delete",
                        { streamPath: item.Path.replace(".flv", "") },
                        x => {
                            if (x == "success") {
                                this.$toast.success("删除成功");
                            } else {
                                this.$toast.error(x);
                            }
                        }
                    );
                }
            });
        },
        toDurationStr(value) {
            if (value > 1000) {
                let s = value / 1000;
                if (s > 60) {
                    s = s | 0;
                    let min = (s / 60) >> 0;
                    if (min > 60) {
                        let hour = (min / 60) >> 0;
                        return hour + "hour" + (min % 60) + "min";
                    } else {
                        return min + "min" + (s % 60) + "s";
                    }
                } else {
                    return s.toFixed(3) + "s";
                }
            } else {
                return value + "ms";
            }
        },
        onVisible(visible) {
            if (visible) {
                this.ajax.getJSON("/record/flv/list", {}, x => {
                    this.data = x;
                });
            }
        }
    }
});

// CONCATENATED MODULE: ./src/components/Records.vue?vue&type=script&lang=js&
 /* harmony default export */ var components_Recordsvue_type_script_lang_js_ = (Recordsvue_type_script_lang_js_); 
// EXTERNAL MODULE: ./src/components/Records.vue?vue&type=style&index=0&id=7c800264&scoped=true&lang=css&
var Recordsvue_type_style_index_0_id_7c800264_scoped_true_lang_css_ = __webpack_require__("0e05");

// CONCATENATED MODULE: ./node_modules/vue-loader/lib/runtime/componentNormalizer.js
/* globals __VUE_SSR_CONTEXT__ */

// IMPORTANT: Do NOT use ES2015 features in this file (except for modules).
// This module is a runtime utility for cleaner component module output and will
// be included in the final webpack user bundle.

function normalizeComponent (
  scriptExports,
  render,
  staticRenderFns,
  functionalTemplate,
  injectStyles,
  scopeId,
  moduleIdentifier, /* server only */
  shadowMode /* vue-cli only */
) {
  // Vue.extend constructor export interop
  var options = typeof scriptExports === 'function'
    ? scriptExports.options
    : scriptExports

  // render functions
  if (render) {
    options.render = render
    options.staticRenderFns = staticRenderFns
    options._compiled = true
  }

  // functional template
  if (functionalTemplate) {
    options.functional = true
  }

  // scopedId
  if (scopeId) {
    options._scopeId = 'data-v-' + scopeId
  }

  var hook
  if (moduleIdentifier) { // server build
    hook = function (context) {
      // 2.3 injection
      context =
        context || // cached call
        (this.$vnode && this.$vnode.ssrContext) || // stateful
        (this.parent && this.parent.$vnode && this.parent.$vnode.ssrContext) // functional
      // 2.2 with runInNewContext: true
      if (!context && typeof __VUE_SSR_CONTEXT__ !== 'undefined') {
        context = __VUE_SSR_CONTEXT__
      }
      // inject component styles
      if (injectStyles) {
        injectStyles.call(this, context)
      }
      // register component module identifier for async chunk inferrence
      if (context && context._registeredComponents) {
        context._registeredComponents.add(moduleIdentifier)
      }
    }
    // used by ssr in case component is cached and beforeCreate
    // never gets called
    options._ssrRegister = hook
  } else if (injectStyles) {
    hook = shadowMode
      ? function () { injectStyles.call(this, this.$root.$options.shadowRoot) }
      : injectStyles
  }

  if (hook) {
    if (options.functional) {
      // for template-only hot-reload because in that case the render fn doesn't
      // go through the normalizer
      options._injectStyles = hook
      // register for functional component in vue file
      var originalRender = options.render
      options.render = function renderWithStyleInjection (h, context) {
        hook.call(context)
        return originalRender(h, context)
      }
    } else {
      // inject component registration as beforeCreate hook
      var existing = options.beforeCreate
      options.beforeCreate = existing
        ? [].concat(existing, hook)
        : [hook]
    }
  }

  return {
    exports: scriptExports,
    options: options
  }
}

// CONCATENATED MODULE: ./src/components/Records.vue






/* normalize component */

var component = normalizeComponent(
  components_Recordsvue_type_script_lang_js_,
  Recordsvue_type_template_id_7c800264_scoped_true_render,
  Recordsvue_type_template_id_7c800264_scoped_true_staticRenderFns,
  false,
  null,
  "7c800264",
  null
  
)

/* harmony default export */ var Records = (component.exports);
// CONCATENATED MODULE: ./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/App.vue?vue&type=script&lang=js&
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//


/* harmony default export */ var Appvue_type_script_lang_js_ = ({
    components: {
        Records: Records
    },
    data() {
        return {
            columns: [
                {
                    title: "房间",
                    name: "StreamPath",
                    sortable: true
                },
                {
                    title: "类型",
                    name: "Type",
                    sortable: true
                },
                {
                    title: "开始时间",
                    name: "StartTime",
                    sortable: true
                },
                {
                    title: "音频格式",
                    name: "AudioInfo"
                },
                {
                    title: "采样率",
                    name: "AudioInfo"
                },
                {
                    title: "声道",
                    name: "AudioInfo"
                },
                {
                    title: "视频格式",
                    name: "VideoInfo"
                },
                {
                    title: "分辨率",
                    name: "VideoInfo"
                },
                {
                    title: "数据包",
                    name: ""
                },
                {
                    title: "订阅者",
                    name: "Subscribes"
                }
            ]
        };
    },
    methods: {
        record(item) {
            let append = false;
            this.$confirm(
                h =>
                    h("mu-switch", {
                        props: {
                            label: "追加模式"
                        },
                        on: {
                            change(value) {
                                append = value;
                            }
                        }
                    }),
                "是否开始录制"
            ).then(result => {
                if (result) {
                    this.ajax.get(
                        "/record/flv?append=" + append,
                        { streamPath: item.StreamPath },
                        x => {
                            if (x == "success") {
                                this.$toast.success(
                                    "开始录制" + (append ? "(追加模式)" : "")
                                );
                            } else {
                                this.$toast.error(x);
                            }
                        }
                    );
                }
            });
        },
        stopRecord(item) {
            this.$confirm("是否停止录制", "提示").then(result => {
                this.ajax.get(
                    "/record/flv/stop",
                    { streamPath: item.StreamPath },
                    x => {
                        if (x == "success") {
                            this.$toast.success("停止录制");
                        } else {
                            this.$toast.error(x);
                        }
                    }
                );
            });
        },
        isRecording(item) {
            return (
                item.SubscriberInfo &&
                item.SubscriberInfo.find(x => x.Type == "FlvRecord")
            );
        }
    }
});

// CONCATENATED MODULE: ./src/App.vue?vue&type=script&lang=js&
 /* harmony default export */ var src_Appvue_type_script_lang_js_ = (Appvue_type_script_lang_js_); 
// EXTERNAL MODULE: ./src/App.vue?vue&type=style&index=0&id=3ce37488&scoped=true&lang=css&
var Appvue_type_style_index_0_id_3ce37488_scoped_true_lang_css_ = __webpack_require__("39a8");

// CONCATENATED MODULE: ./src/App.vue






/* normalize component */

var App_component = normalizeComponent(
  src_Appvue_type_script_lang_js_,
  render,
  staticRenderFns,
  false,
  null,
  "3ce37488",
  null
  
)

/* harmony default export */ var App = (App_component.exports);
// CONCATENATED MODULE: ./node_modules/@vue/cli-service/lib/commands/build/entry-lib.js


/* harmony default export */ var entry_lib = __webpack_exports__["default"] = (App);



/***/ })

/******/ })["default"];
});
//# sourceMappingURL=plugin-record.umd.js.map