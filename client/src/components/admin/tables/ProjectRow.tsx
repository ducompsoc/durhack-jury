import {act, FocusEvent, useEffect, useRef, useState} from 'react';
import {errorAlert, timeSince} from '../../../util';
import HidePopup from './HidePopup';
import InfoPopup from './InfoPopup';
import DeletePopup from './DeletePopup';
import EditProjectPopup from './EditProjectPopup';
import useAdminStore from '../../../store';
import {postRequest} from '../../../api';
import {Project, YesNoResponse} from "../../../types";

interface ProjectRowProps {
    project: Project;
    idx: number;
    checked: boolean;
    handleCheckedChange: (e: React.ChangeEvent<HTMLInputElement>, idx: number) => void;
}

const ProjectRow = ({project, idx, checked, handleCheckedChange}: ProjectRowProps) => {
    const [popup, setPopup] = useState(false);
    const [editPopup, setEditPopup] = useState(false);
    const [deletePopup, setDeletePopup] = useState(false);
    const [hidePopup, setHidePopup] = useState(false);
    const [infoPopup, setInfoPopup] = useState(false);
    const ref = useRef<HTMLDivElement>(null);
    const fetchProjects = useAdminStore((state) => state.fetchProjects);

    useEffect(() => {
        function closeClick(event: MouseEvent) {
            if (ref && ref.current && !ref.current.contains(event.target as Node)) {
                setPopup(false);
            }
        }

        // Bind the event listener
        document.addEventListener('mousedown', closeClick);
        return () => {
            // Unbind the event listener on clean up
            document.removeEventListener('mousedown', closeClick);
        };
    }, [ref]);

    const doAction = async (action: 'edit' | 'prioritize' | 'hide' | 'un-hide' | 'info' | 'delete') => {
        console.log('Performing action: ' + action + ' on project ' + project.name);
        switch (action) { // todo: find a way to reuse types for this functon for assignments
            case 'edit':
                // Open edit popup
                setEditPopup(true);
                break;
            case 'hide':
                // Open hide popup
                setHidePopup(true);
                break;
            case 'un-hide':
                // Un-hide project
                unhideProject();
                break;
            case 'info':
                // Open info popup
                setInfoPopup(true);
                break;
            case 'delete':
                // Open delete popup
                setDeletePopup(true);
                break;
        }

        setPopup(false);
    };

    const unhideProject = async () => {
        const res = await postRequest<YesNoResponse>('/project/unhide', {id: project.id});
        if (res.status === 200) {
            alert(`Project un-hidden successfully!`);
            await fetchProjects();
        } else {
            errorAlert(res);
        }
    };

    const onInputFocusLoss = async (e: FocusEvent<HTMLInputElement>) => {
        const res = await postRequest<YesNoResponse>('/project/update-location', {id: project.id, location: e.target.value});
        if (res.status === 200) {
            fetchProjects();
        } else {
            errorAlert(res);
        }
    }

    return (
        <>
            {/*todo: highlight projects that are repeatedly (can be a variable) flagged as absent*/}
            <tr
                key={idx}
                className={
                    'border-t-2 border-backgroundDark duration-150 ' +
                    (checked
                        ? 'bg-primary/20'
                        : !project.active
                        ? 'bg-lightest'
                        : 'bg-background')
                }
            >
                <td className="px-2">
                    <input
                        type="checkbox"
                        checked={checked || false}
                        onChange={(e) => {
                            handleCheckedChange(e, idx);
                        }}
                        className="cursor-pointer hover:text-primary duration-100"
                    ></input>
                </td>
                <td className="[&:not(:hover)]:truncate hover:break-words hover:text-wrap">{project.name}</td>
                <td className="text-center">{project.guild}</td>
                <td className="text-center py-1">
                    <input
                        className="w-full md:w-2/3 rounded-2xl"
                        name="location"
                        key={project.id}
                        defaultValue={project.location}
                        type="text"
                        onBlur={(e) => onInputFocusLoss(e)}
                    />
                </td>
                <td className="text-center">{project.score} [{project.seen > 0 ? (project.score/project.seen).toFixed(2) : project.score}]</td>
                <td className="text-center">{project.seen}</td>
                <td className="text-center">{timeSince(project.last_activity)}</td>
                <td className="text-right font-bold flex align-center justify-end">
                    {popup &&
                        <div
                            className="absolute flex flex-col bg-background rounded-md border-lightest border-2 font-normal text-sm"
                            ref={ref}
                        >
                            {['Info', project.active ? 'Hide' : 'Un-hide', 'Delete'].map((action) => 
                                <div
                                    key={action}
                                    className={`py-1 pl-4 pr-2 cursor-pointer hover:bg-primary/20 duration-150 ${action == 'Delete' ? 'text-error' : ''}`}
                                    onClick={() => { doAction(action.toLowerCase() as 'edit' | 'prioritize' | 'hide' | 'un-hide' | 'info' | 'delete'); }}
                                >
                                    {action}
                                </div>
                            )}
                        </div>
                    } 
                    <span
                        className="cursor-pointer px-1 hover:text-primary duration-150"
                        onClick={() => {
                            setPopup(!popup);
                        }}
                    >
                        ...
                    </span>
                </td>
            </tr>
            {deletePopup && <DeletePopup element={project} close={setDeletePopup} />}
            {hidePopup && <HidePopup projects={[project]} close={setHidePopup} />}
            {infoPopup && <InfoPopup project={project} close={setInfoPopup} />}
        </>
    );
};
 // ['Info', project.active ? 'Hide' : 'Un-hide', 'Delete'].map((str) => (
export default ProjectRow;
