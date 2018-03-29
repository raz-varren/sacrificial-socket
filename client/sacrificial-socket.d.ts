//sacrificial-socket.d.ts is the type definition file for sacrificial-socket.js
//
//because the sacrificial-socket client is just a class constructor
//the way to import this into TypeScript is with:
//	import SS = require('sacrificial-socket');

declare namespace SS {
        export interface connOpts{
                enabled?: boolean;
                replayOnConnect?: boolean;
                intervalMS?: number;
        }

	export interface ssOpts{
                reconnectOpts?: connOpts;
        }
}


declare class SS{
        constructor(url: string, opts?: SS.ssOpts);

        noop(): void;

        onConnect(callback: (event: any) => void): void;
        onDisconnect(callback: (event: any) => void): void;

        on(eventName: string, callback: (data: any) => void): void;
        off(eventName: string): void;

        emit(eventName: string, data: any): void;

        close(): any;
}


export = SS;
