(function(window){ 'use strict';
	if(!window.angular){
		throw 'angular not loaded';
		return;
	}
	
	if(!window.SS){
		throw 'sacrificial socket not loaded';
		return;
	}
	
	window.angular.module('sacrificial-socket', [])
		.factory('ss', ['$window', '$rootScope', '$log', function($window, $rootScope, $log){
			function SSNG(url, opts){
				var self = this,
					socket = new $window.SS(url, opts);
					
				self.onConnect = function(callback){
					callback = callback || socket.noop;
					socket.onConnect(function(){
						var args = arguments;
						$rootScope.$apply(function(){
							callback.apply(self, args);
						})
					});
				};
				
				self.onDisconnect = function(callback){
					callback = callback || socket.noop;
					socket.onDisconnect(function(){
						var args = arguments;
						$rootScope.$apply(function(){
							callback.apply(self, args);
						});
					});
				};
				
				self.on = function(eventName, callback){
					callback = callback || socket.noop;
					socket.on(eventName, function(){
						var args = arguments;
						$rootScope.$apply(function(){
							callback.apply(self, args);
						});
					});
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