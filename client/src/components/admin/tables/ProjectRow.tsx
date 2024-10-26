import {FocusEvent, useEffect, useRef, useState} from 'react';
import {errorAlert, timeSince} from '../../../util';
import DeletePopup from './DeletePopup';
import EditProjectPopup from './EditProjectPopup';
import useAdminStore from '../../../store';
import {postRequest} from '../../../api';

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

    const doAction = (action: 'edit' | 'prioritize' | 'hide' | 'delete') => {
        switch (action) {
            case 'edit':
                // Open edit popup
                setEditPopup(true);
                break;
            case 'hide':
                // Hide
                hideProject();
                break;
            case 'delete':
                // Open delete popup
                setDeletePopup(true);
                break;
        }

        setPopup(false);
    };

    const hideProject = async () => {
        const res = await postRequest<YesNoResponse>(project.active ? '/project/hide' : '/project/unhide', {id: project.id});
        if (res.status === 200) {
            alert(`Project ${project.active ? 'hidden' : 'un-hidden'} successfully!`);
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
                        checked={checked}
                        onChange={(e) => {
                            handleCheckedChange(e, idx);
                        }}
                        className="cursor-pointer hover:text-primary duration-100"
                    ></input>
                </td>
                <td className="[&:not(:hover)]:truncate hover:break-words hover:text-wrap">{project.name}</td>
                <td className="text-center py-1">
                    <input
                        className="w-full md:w-2/3"
                        name="location"
                        key={project.id}
                        defaultValue={project.location}
                        type="text"
                        onBlur={(e) => onInputFocusLoss(e)}
                    />
                </td>
                <td className="text-center">{project.score}</td>
                <td className="text-center">{project.seen}</td>
                <td className="text-center">{timeSince(project.last_activity)}</td>
                <td className="text-right font-bold flex align-center justify-end">
                    {popup && (
                        <div
                            className="absolute flex flex-col bg-background rounded-md border-lightest border-2 font-normal text-sm"
                            ref={ref}
                        >
                            <div
                                className="py-1 pl-4 pr-2 cursor-pointer hover:bg-primary/20 duration-150"
                                onClick={() => doAction('hide')}
                            >
                                {project.active ? 'Hide' : 'Un-hide'}
                            </div>
                            <div
                                className="py-1 pl-4 pr-2 cursor-pointer hover:bg-primary/20 duration-150 text-error"
                                onClick={() => doAction('delete')}
                            >
                                Delete
                            </div>
                        </div>
                    )}
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
            {editPopup && <EditProjectPopup project={project} close={setEditPopup} />}
        </>
    );
};

export default ProjectRow;
