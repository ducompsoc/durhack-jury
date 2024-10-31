import { useState } from 'react';
import Button from '../Button';
import FlagsPopup from './FlagsPopup';

const AdminToolbar = (props: { showProjects: boolean }) => {
    const [showFlags, setShowFlags] = useState(false);
    return (
        <div className="flex flex-row px-8 py-4 items-center">
            <div>
                {props.showProjects && (
                    <Button
                        type="outline"
                        square
                        bold
                        full
                        className="py-2 px-4 rounded-md"
                        href='/admin/add-projects'
                    >
                        Add Projects
                    </Button>
                )}
            </div>
            <div className="ml-4">
                {props.showProjects && (
                    <Button
                        type="outline"
                        square
                        bold
                        full
                        className="py-2 px-4 rounded-md"
                        onClick={() => {
                            setShowFlags(true);
                        }}
                    >
                        See Flags
                    </Button>
                )}
            </div>
            <div className="ml-4 italic">
                Click on headings to sort by that column. This will clear your selections.
            </div>
            {showFlags && <FlagsPopup close={setShowFlags}/>}
        </div>
    );
};

export default AdminToolbar;
