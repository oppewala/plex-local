import React, {useCallback, useContext, useEffect, useMemo, useState} from "react";

interface Message {
    MessageType: string
}

interface DownloadUpdateMessage extends Message {
    Title: string
    BytesDownloaded: number
    TotalBytes: number
}

type Download = {
    Title: string
    BytesDownloaded: number
    TotalBytes: number
    Complete: boolean
}

export const Downloads = () => {
    const [downloads, setDownloads] = useState<{ [title: string]: Download }>({
        'Lord': {
            Title: 'Lord',
            BytesDownloaded: 10000,
            TotalBytes: 100000,
            Complete: false
        }
    })

    const handleDownloadMessage = (message: Message) => {
        const msg = message as DownloadUpdateMessage;
        setDownloads((prevState: { [title: string]: Download }) => {
            let newState = {
                ...prevState
            };
            newState[msg.Title] = {
                Title: msg.Title,
                BytesDownloaded: msg.BytesDownloaded,
                TotalBytes: msg.TotalBytes,
                Complete: false
            };
            return newState;
        });
    }

    const handleCompleteMessage = (message: Message) => {
        const msg = message as DownloadUpdateMessage;
        setDownloads((prevState: { [title: string]: Download }) => {
            let newState = {
                ...prevState
            };
            newState[msg.Title] = {
                Title: msg.Title,
                BytesDownloaded: 100,
                TotalBytes: 100,
                Complete: true
            };
            return newState;
        });
    }

    useMemo(() => {
        const ws = new WebSocket('ws://localhost:8080/api/ws')

        ws.addEventListener('open', ev => {
            console.log('open', ev)
        })
        ws.addEventListener('message', ev => {
            if (ev.data === 'PING') {
                return;
            }

            const message: Message = JSON.parse(ev.data)
            if (message.MessageType === "download-update") {
                handleDownloadMessage(message);
                return;
            }
            if (message.MessageType === 'download-complete') {
                handleCompleteMessage(message);
            }

            console.log('Unhandled type', message.MessageType, message)
        })
        ws.addEventListener('close', ev => {
            // Reconnect
            console.log('closed!', ev)
        })

        return ws;
    }, []);

    return (
        <div className='w-full'>
            <ul className='w-full'>{Object.getOwnPropertyNames(downloads).map((t) => <DownloadBar key={t} title={t} downloads={downloads} />)}</ul>
        </div>
    )
}

const DownloadBar :React.FC<{title: string, downloads: { [title: string]: Download }}> = ({ title, downloads }) => {
    const download = downloads[title];
    const perc = download.BytesDownloaded / download.TotalBytes * 100
    const progressStyle = {
        width: perc + '%'
    }

    return (
        <li className='relative grid-cols-1 text-md my-6 overflow-hidden rounded-xl shadow-md px-6 py-3'>
            <div>
                <div className='-z-10 absolute inset-0 w-full h-full bg-green-100'></div>
                <div className={`-z-10 absolute inset-0 h-full bg-gradient-to-r from-green-400 to-green-500 ${!download.Complete ? 'animate-pulse' : ''}`} style={progressStyle}></div>
                { !download.Complete && <div className={`-z-10 absolute inset-0 h-full bg-none border-solid border-r border-green-600 ${!download.Complete ? 'animate-pulse' : ''}`} style={progressStyle}></div>}
            </div>
            <div className=''>{title} ({Math.round((perc + Number.EPSILON) * 10) / 10}%)</div>
        </li>)
}

export default Downloads;