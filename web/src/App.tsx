import React from 'react';
import {Search} from '@components/Search';
import Downloads from "@components/Downloads/Downloads";


function App() {
    return (
        <div className="p-6">
            <div className='flex flex-wrap mx-auto max-w-3xl'>
                <Downloads />
                <Search />
            </div>
        </div>
    );
}

export default App;
