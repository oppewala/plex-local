import React from "react";

export interface Message {
    MessageType: string;
}

interface Listener {
    Type: string;
    Emit: (ev: Event) => void;
}

interface MessageListener {
    MessageType: string;
    Emit: (msg: Message) => void;
}

interface ISocketContext {
    Connection: WebSocket;
    Listeners: Array<Listener | MessageListener>
    AddListener: (l: Listener | MessageListener) => void;
}

class SocketCtx implements ISocketContext {
    Connection: WebSocket;
    Listeners: Array<Listener | MessageListener>;

    constructor(conn: WebSocket) {
        this.Listeners = [];

        this.Connection = conn;
        this.Connection.addEventListener('open', ev => this.emitOpen(ev));
        this.Connection.addEventListener('close', ev => this.emitClose(ev));
        this.Connection.addEventListener('message', ev => this.emitMessage(ev));
        this.Connection.addEventListener('error', ev => this.emitError(ev));
    }

    private emitOpen(ev: Event) {
        for (const listener of this.Listeners) {
            if (!SocketCtx.isMessageListener(listener) && listener.Type === 'open') {
                listener.Emit(ev);
            }
        }
    }

    private emitClose(ev: CloseEvent) {
        for (const listener of this.Listeners) {
            if (!SocketCtx.isMessageListener(listener) && listener.Type === 'close') {
                listener.Emit(ev);
            }
        }
    }

    private emitMessage(ev: MessageEvent) {
        if (ev.data === 'PING') {
            // console.log('[WS] PING', ev)
            return;
        }

        const message: Message = JSON.parse(ev.data);
        for (const listener of this.Listeners) {
            if (SocketCtx.isMessageListener(listener) &&listener.MessageType === message.MessageType) {
                listener.Emit(message);
            }
        }
    }

    private emitError(ev: Event) {
        for (const listener of this.Listeners) {
            if (!SocketCtx.isMessageListener(listener) && listener.Type === 'error') {
                listener.Emit(ev);
            }
        }
    }

    AddListener(l: Listener | MessageListener): void {
        this.Listeners.push(l);
    }

    private static isMessageListener(l: Listener | MessageListener): l is MessageListener {
        return (l as MessageListener).MessageType !== undefined;
    }
}

const ws = new WebSocket('ws://localhost:8080/api/ws');
const defaultContext = new SocketCtx(ws);

const SocketContext = React.createContext<ISocketContext>(defaultContext)

export default SocketContext