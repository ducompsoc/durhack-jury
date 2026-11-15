import { errorAlert, convertUnixTimestamp } from "../../../util";
import { postRequest } from "../../../api";
import useAdminStore from "../../../store";
import { HiddenReason, HiddenReasonWithCount, Project, YesNoResponse } from "../../../types";
import { useState } from "react";

// sort hidden reasons by frequency
const getHiddenReasons = (projects: Project[]): HiddenReasonWithCount[] => {
    const frequency: { [reason: string]: number } = {};
    projects.forEach(project => {
        project.hidden_reasons.forEach(reason => {
            let jsonReason = JSON.stringify(reason);
            frequency[jsonReason] = (frequency[jsonReason] || 0) + 1
        });
    });
    return Object.entries(frequency).sort((a, b) => b[1] - a[1]).map(([reason, count]) => ({
        ...JSON.parse(reason),
        count: count
    }));
}

interface UnhidePopupProps {
    /* Projects to unhide */
    projects: Project[];

    /* Function to modify the popup state variable */
    close: React.Dispatch<React.SetStateAction<boolean>>;
}

const UnhidePopup = ({ projects, close }: UnhidePopupProps) => {
    const [selectedReasonIdx, setSelectedReasonIdx] = useState(-1);
    const fetchProjects = useAdminStore((state) => state.fetchProjects);
    const hiddenReasons = getHiddenReasons(projects);
    
    const unhideProjects = async () => {
        if (selectedReasonIdx == -1) {
            alert('Please select a reason.');
            return;
        }
        const res = await postRequest<YesNoResponse>('/project/unhide-many', {
            ids: projects.map(project => project.id),
            hidden_reason: hiddenReasons[selectedReasonIdx],
        });
        if (res.status === 200) {
            alert(`Project un-hidden successfully!`);
            await fetchProjects();
        } else {
            errorAlert(res);
        }
        close(false);
    };


    return (
        <>
            <div className="bg-background fixed z-20 left-1/2 top-1/2 translate-x-[-50%] translate-y-[-50%] py-6 px-10 w-1/3">
                <h1 className="text-5xl font-bold mb-2 text-center">Select a hidden reason to remove</h1>
                <div className="flex-col gap-2">
                    {hiddenReasons.map((entry, idx) =>
                        <div
                            key={entry.when}
                            className={`space-y-0.5 p-2 rounded-md ${selectedReasonIdx == idx ? 'bg-lightest' : 'bg-primary/20'}`}
                            onClick={() => setSelectedReasonIdx(idx)}
                        >
                            <p className="text-lg font-semibold text-black">{entry.reason}</p>
                            <div className="flex gap-2">
                                <p className="font-medium text-gray-500 flex-1">{convertUnixTimestamp(entry.when)}</p>
                                <p className="font-medium text-gray-500 flex-1">Occurences: {entry.count}</p>
                            </div>
                        </div>
                    )}
                </div>
                <div className="flex flex-row justify-around">
                    <button
                        className="border-lightest border-2 rounded-full px-6 py-1 mt-4 w-2/5 font-bold text-2xl text-lighter hover:bg-lighter/30 duration-200"
                        onClick={() => close(false)}
                    >
                        Cancel
                    </button>
                    <button
                        className="bg-primary rounded-full px-6 py-1 mt-4 w-2/5 font-bold text-2xl text-background hover:bg-primary/80 duration-200"
                        onClick={unhideProjects}
                    >
                        Submit
                    </button>
                </div>
            </div>
        </>
    )
}

export default UnhidePopup;