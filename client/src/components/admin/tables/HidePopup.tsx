import { postRequest } from "../../../api";
import { Project } from "../../../types";
import { useState } from "react";
import { errorAlert } from "../../../util";
import useAdminStore from "../../../store";

interface HideProjectPopupProps {
    /* Projects to edit */
    projects: Project[];

    /* Function to modify the popup state variable */
    close: React.Dispatch<React.SetStateAction<boolean>>;
}

const HideProjectPopup = ({ projects, close }: HideProjectPopupProps) => {
    const [selectedOption, setSelectedOption] = useState('');
    const [selectedReason, setSelectedReason] = useState('');
    const fetchProjects = useAdminStore((state) => state.fetchProjects);
    const options = ['Lunch', 'Not found', 'Option 3', 'Other'];

    const hideProject = async () => {
        if (selectedReason == '' || selectedReason == 'Other') {
            alert('Please select a reason for hiding the project(s).');
            return;
        }
        const res = await postRequest('/project/hide-many', 
            { 
                ids: projects.map(project => project.id), 
                reason: selectedReason 
            }
        );
        if (res.status === 200) {
            alert(`Project hidden successfully!`);
        } else {
            errorAlert(res);
        }
        await fetchProjects();
        close(false);
    }
    return (
        <>
            <div className="bg-background fixed z-20 left-1/2 top-1/2 translate-x-[-50%] translate-y-[-50%] py-6 px-10 w-1/3">
                <h1 className="text-5xl font-bold mb-2 text-center">Enter a reason for hiding the project(s)</h1>
                <div className="flex flex-row justify-around mt-4">
                    <ul>
                        {options.map((option) => 
                            <div
                                key={option}
                                className={`max-w-sm rounded-md p-4 mg-12 gap-4 ${selectedOption == option ? 'bg-lightest' : 'bg-primary/20'}`}
                                onClick={() => { 
                                    setSelectedOption(option); 
                                    setSelectedReason(option); 
                                }}
                            >
                                {option}
                            </div>
                        )}
                    </ul>
                </div>
                <div className="bg-background flex flex-row mt-4">
                    {selectedOption == 'Other' && 
                    <textarea 
                        className="border-lightest border-2 rounded-md p-2 resize-none w-full" 
                        placeholder="Enter reason here..." 
                        onChange={(e) => { setSelectedReason(e.target.value); }}
                    />}
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
                        onClick={hideProject}
                    >
                        Submit
                    </button>
                </div>
            </div>
        </>
    );
};

export default HideProjectPopup;
