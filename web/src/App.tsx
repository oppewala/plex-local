import React from 'react';
import {Search} from '@components/Search';
import Downloads from "@components/Downloads/Downloads";


function App() {
    return (
        <div className="p-6">
            <header className="flex mx-auto text-3xl max-w-3xl mt-6 mb-12">
                <h1>Plex Local DL</h1>
            </header>
            <div className='flex flex-wrap mx-auto max-w-3xl'>
                <Downloads />
                <Search />
            </div>
        </div>
    );
}

export default App;
