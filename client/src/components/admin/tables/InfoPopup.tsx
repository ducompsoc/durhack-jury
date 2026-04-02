import { useState } from 'react';
import { HiddenReason, Project } from '../../../types';
import { convertUnixTimestamp } from '../../../util';

interface InfoPopupProps {
    /* Project to show */
    project: Project;

    /* Function to modify the popup state variable */
    close: React.Dispatch<React.SetStateAction<boolean>>;
}

const renderHiddenReasonRow = (hiddenReason: HiddenReason) => (
    <li className="font-mono" key={hiddenReason.when}>{convertUnixTimestamp(hiddenReason.when)}: {hiddenReason.reason}</li>
);

const InfoPopup = ({ project, close }: InfoPopupProps) => {
    const [showAll, setShowAll] = useState(false);
    return (
        <>
            <div
                className="fixed left-0 top-0 z-20 w-screen h-screen bg-black/30"
                onClick={() => close(false)}
            ></div>
            <div className="bg-background fixed z-30 left-1/2 top-1/2 translate-x-[-50%] translate-y-[-50%] py-6 px-10 w-1/3">
                <h1 className="text-5xl font-bold mb-2 text-center">About <span className="text-primary">{project.name}</span></h1>
                <p className="text-xl"><b>Description:</b> {project.description}</p>
                <p className="text-xl"><b>Guild:</b> {project.guild}</p>
                <p className="text-xl"><b>Location:</b> {project.location}</p>
                <p className="text-xl"><b>Score:</b> {project.score}</p>
                <p className="text-xl"><b>Status:</b> {project.active ? 'Active' : 'Hidden'}</p>
                {project.hidden_reasons.length > 0 && (
                    <div className="mt-4 rounded-lg border border-gray-300 bg-white/80 p-4">
                        <p className="mb-2 text-xl font-semibold text-gray-800">Hide reasons:</p>
                        <ul className="m-0 p-0">
                            {showAll ? 
                                project.hidden_reasons.toReversed().map(hiddenReason => renderHiddenReasonRow(hiddenReason)) :
                                renderHiddenReasonRow(project.hidden_reasons[project.hidden_reasons.length - 1])
                            }
                        </ul>
                        <span
                            className="mt-2 inline-block cursor-pointer text-pink-500 hover:text-pink-700 transition-colors"
                            onClick={() => setShowAll(!showAll)}
                        >
                            {showAll ? 'Hide' : 'Show more'}
                        </span>
                    </div>
                )}
                <div className="flex flex-row justify-around">
                    <button
                        className=" border-lightest border-2 rounded-full px-6 py-1 mt-4 w-2/5 font-bold text-2xl text-lighter hover:bg-lighter/30 duration-200"
                        onClick={() => close(false)}
                    >
                        Close
                    </button>
                </div>
            </div>
        </>
    );
};

export default InfoPopup;
