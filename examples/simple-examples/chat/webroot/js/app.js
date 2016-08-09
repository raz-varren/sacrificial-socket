(function(SS){ 'use strict';
	var msgCnt      = document.querySelector('#messages'),
		msgList     = document.querySelector('#message-list'),
		msgClear    = document.querySelector('#clear-messages-btn'),
		nameInput   = document.querySelector('#name-input'),
		joinInput   = document.querySelector('#join-room-input'),
		joinBtn     = document.querySelector('#join-room-btn'),
		msgBody     = document.querySelector('#message-body-input'),
		send        = document.querySelector('#message-send-btn'),
		currentRoom = null,
		months      = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
		
		getTime = function(dt) {
			var hours = dt.getHours(),
				minutes = dt.getMinutes(),
				ampm = hours >= 12 ? 'pm' : 'am';
			hours %= 12;
			hours = hours ? hours : 12; // the hour '0' should be '12'
			minutes = minutes < 10 ? '0'+minutes : minutes;
			var strTime = hours + ':' + minutes + ' ' + ampm;
			return strTime;
		},
		addMsg = function(msg){
			var li = document.createElement('li'),
				dt = new Date(),
				dtString = months[dt.getMonth()]+' '+dt.getDate()+', '+dt.getFullYear();
			li.innerText = dtString+' '+getTime(dt)+' - '+msg;
			msgList.appendChild(li);
			msgCnt.scrollTop = msgCnt.scrollHeight;
		};
	
	
	var ss = new SS('ws://'+window.location.host+'/socket');
	
	ss.onConnect(function(){
		ss.emit('join', joinInput.value);
	});
	
	ss.onDisconnect(function(){
		alert('chat disconnected');
	});
	
	ss.on('joinedRoom', function(room){
		currentRoom = room;
		addMsg('joined room: '+room);
	});
	
	ss.on('message', function(msg){
		addMsg(msg);
	});
	
	send.addEventListener('click', function(){
		var msg = msgBody.value;
		if(msg.length === 0) return;
		
		ss.emit('message', {Room: currentRoom, Message: nameInput.value+' says: '+msg});
	});
	
	joinBtn.addEventListener('click', function(){
		var room = joinInput.value;
		if(room.length === 0) return;
		ss.emit('join', joinInput.value);
	});
	
	msgClear.addEventListener('click', function(){
		msgList.innerHTML = '';
	});
	
})(window.SS);