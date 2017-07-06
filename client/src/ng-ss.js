(function(window){ 'use strict';
	if(!window.angular){
		throw 'angular not loaded';
	}
	
	if(!window.SS){
		throw 'sacrificial socket not loaded';
	}
	
	window.angular.module('sacrificial-socket', [])
		.factory('ss', ['$window', '$rootScope', function($window, $rootScope){
			function SSNG(url, opts){
				var self = this,
					socket = new $window.SS(url, opts),
					scopeCB = function(cb){
						return function(){
							var args = arguments;
							$rootScope.$apply(function(){
								cb.apply(self, args);
							});
						};
					};
				
				self.onConnect = function(callback){
					socket.onConnect(scopeCB(callback));
				};
				
				self.onDisconnect = function(callback){
					socket.onDisconnect(scopeCB(callback));
				};
				
				self.on = function(eventName, callback){
					socket.on(eventName, scopeCB(callback));
				};
				
				self.off = function(eventName){
					return socket.off(eventName);
				};
				
				self.emit = function(eventName, data){
					return socket.emit(eventName, data);
				};
				
				self.close = function(){
					return socket.close();
				};
			}
			
			return function(url, opts){ return new SSNG(url, opts); };
		}]);
})(window);