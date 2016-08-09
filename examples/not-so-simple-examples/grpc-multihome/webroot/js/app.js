(function(window){ 'use strict';
	var ws = new window.SS((window.location.protocol === 'https:' ? 'wss':'ws')+'://'+window.location.host+'/socket'),
	get = function(selector){
		return document.querySelector(selector);
	},
	rand = function(min, max){
		return Math.floor(Math.random() * (max - min + 1) + min);
	},
	abToStr = function(ab){
		return String.fromCharCode.apply(null, new Uint8Array(ab));
	},
	strToAB = function(str) {
		var buf = new ArrayBuffer(str.length);
		var dv = new DataView(buf);
		for (var i=0, strLen=str.length; i<strLen; i++) {
			dv.setUint8(i, str.charCodeAt(i));
		}
		return buf;
	},
	wierdness         = null,
	inEcho            = get('#in-echo'),
	inJoin            = get('#in-join'),
	inLeave           = get('#in-leave'),
	inBroadcast       = get('#in-broadcast'),
	inRoomcastRoom    = get('#in-roomcast-room'),
	inRoomcastData    = get('#in-roomcast-data'),
	btnEcho           = get('#btn-echo'),
	btnEchoBin        = get('#btn-echo-bin'),
	btnEchoJSON       = get('#btn-echo-json'),
	btnJoin           = get('#btn-join'),
	btnLeave          = get('#btn-leave'),
	btnBroadcast      = get('#btn-broadcast'),
	btnBroadcastBin   = get('#btn-broadcast-bin'),
	btnBroadcastJSON  = get('#btn-broadcast-json'),
	btnRoomcast       = get('#btn-roomcast'),
	btnRoomcastBin    = get('#btn-roomcast-bin'),
	btnRoomcastJSON   = get('#btn-roomcast-json'),
	btnClose          = get('#btn-close'),
	btnClear          = get('#btn-clear'),
	btnGetWierd       = get('#btn-wierd'),
	btnGetNormal      = get('#btn-normal'),
	messages          = get('#messages'),
	addMessage        = function(msg){
		var li = document.createElement('li'),
			dt = new Date(),
			li = document.createElement('li');
		
		li.innerText = msg;
		messages.appendChild(li);
		messages.scrollTop = messages.scrollHeight;
	};
	
	ws.onConnect(function(){
		addMessage('ready');
		ws.emit('echo', 'test ping');
	});
	
	ws.onDisconnect(function(){
		addMessage('disconnected');
	});
	
	ws.on('echo', function(data){
		addMessage(data);
	});
	
	ws.on('echobin', function(data){
		addMessage('got binary: '+data.byteLength+' bytes - '+abToStr(data));
	});
	
	ws.on('echojson', function(data){
		addMessage('got JSON: '+JSON.stringify(data));
	});
	
	ws.on('roomcast', function(data){
		addMessage('got roomcast: '+data);
	});
	
	ws.on('roomcastbin', function(data){
		addMessage('got binary roomcast: '+data.byteLength+' bytes - '+abToStr(data));
	});
	
	ws.on('roomcastjson', function(data){
		addMessage('got JSON roomcast: '+JSON.stringify(data));
	});
	
	ws.on('broadcast', function(data){
		addMessage('got broadcast: '+data);
	});
	
	ws.on('broadcastbin', function(data){
		addMessage('got binary broadcast: '+data.byteLength+' bytes - '+abToStr(data));
	});
	
	ws.on('broadcastjson', function(data){
		addMessage('got JSON broadcast: '+JSON.stringify(data));
	});
	
	btnGetWierd.addEventListener('click', function(){
		getWierd();
	});
	
	btnGetNormal.addEventListener('click', function(){
		if(wierdness){
			window.clearInterval(wierdness);
			wierdness = null;
		}
	});
	
	btnEcho.addEventListener('click', function(){
		if(inEcho.value.length === 0) return;
		ws.emit('echo', inEcho.value);
	});
	
	btnEchoBin.addEventListener('click', function(){
		if(inEcho.value.length === 0) return;
		ws.emit('echobin', strToAB(inEcho.value));
	});
	
	btnEchoJSON.addEventListener('click', function(){
		if(inEcho.value.length === 0) return;
		ws.emit('echojson', {message: inEcho.value});
	});
	
	btnJoin.addEventListener('click', function(){
		if(inJoin.value.length === 0) return;
		ws.emit('join', inJoin.value);
	});
	
	btnLeave.addEventListener('click', function(){
		if(inLeave.value.length === 0) return;
		ws.emit('leave', inLeave.value);
	});
	
	btnBroadcast.addEventListener('click', function(){
		if(inBroadcast.value.length === 0) return;
		ws.emit('broadcast', inBroadcast.value);
	});
	
	btnBroadcastBin.addEventListener('click', function(){
		if(inBroadcast.value.length === 0) return;
		ws.emit('broadcastbin', strToAB(inBroadcast.value));
	});
	
	btnBroadcastJSON.addEventListener('click', function(){
		if(inBroadcast.value.length === 0) return;
		ws.emit('broadcastjson', {message: inBroadcast.value});
	});
	
	btnRoomcast.addEventListener('click', function(){
		if(inRoomcastRoom.value.length === 0 || inRoomcastData.value.length === 0) return;
		ws.emit('roomcast', JSON.stringify({room: inRoomcastRoom.value, data: inRoomcastData.value}));
	});
	
	btnRoomcastBin.addEventListener('click', function(){
		if(inRoomcastRoom.value.length === 0 || inRoomcastData.value.length === 0) return;
		ws.emit('roomcastbin', strToAB(JSON.stringify({room: inRoomcastRoom.value, data: inRoomcastData.value})));
	});
	
	btnRoomcastJSON.addEventListener('click', function(){
		if(inRoomcastRoom.value.length === 0 || inRoomcastData.value.length === 0) return;
		ws.emit('roomcastjson', {room: inRoomcastRoom.value, data: inRoomcastData.value});
	});
	
	btnClose.addEventListener('click', function(){
		ws.close();
	});
	
	btnClear.addEventListener('click', function(){
		messages.innerHTML = "";
	});
	
	function getWierd(){
		var rooms = [
			'unknownplace',
			'trl',
			'purgatory',
			'southdakota',
			'animalhouse',
			'orangecounty',
			'andyrichtersbasement',
			'baldpersonemporium'
		],
		phrases = [
			'welcome to hell',
			'alls wells thats ends wells',
			'we\'ve been waiting for you',
			'did you ever stop to think how that makes your insurance adjuster feel?',
			'with these weaponized puppies I will finally rule the world',
			'my only friend is this series of ones and zeros',
			'how much wood could a woodchuck chuck if a woodchuck was really drunk',
			'if it\'s not syphillis then why does it itch so much',
		],
		actions = [
			'echo',
			'echobin',
			'echojson',
			'join',
			'leave',
			'broadcast',
			'broadcastbin',
			'broadcastjson',
			'roomcast',
			'roomcastbin',
			'roomcastjson'
		],
		i = 1;
		
		wierdness = setInterval(function(){
			//if(ws.readyState !== 1) return;
			var action = actions[rand(0, actions.length-1)],
				phrase = phrases[rand(0, phrases.length-1)],
				room   = rooms[rand(0, rooms.length-1)];
				phrase = (Date.now()/1000)+' - '+phrase;
				
			if(action == 'echo'){
				ws.emit('echo', phrase);
			}else if(action == 'echobin'){
				ws.emit('echobin', strToAB(phrase));
			}else if(action == 'echojson'){
				ws.emit('echojson', {message: phrase});
			}else if(action == 'join'){
				ws.emit('join', room);
			}else if(action == 'leave'){
				ws.emit('leave', room);
			}else if(action == 'broadcast'){
				ws.emit('broadcast', phrase);
			}else if(action == 'broadcastbin'){
				ws.emit('broadcastbin', strToAB(phrase));
			}else if(action == 'broadcastjson'){
				ws.emit('broadcastjson', {message: phrase});
			}else if(action == 'roomcast'){
				ws.emit('roomcast', JSON.stringify({room: room, data: phrase}));
			}else if(action == 'roomcastbin'){
				ws.emit('roomcastbin', strToAB(JSON.stringify({room: room, data: phrase})));
			}else if(action == 'roomcastjson'){
				ws.emit('roomcastjson', {room: room, data: phrase})
			}
		}, 100);
	}
})(window);






















