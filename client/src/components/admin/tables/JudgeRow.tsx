import {useEffect, useRef, useState} from 'react';
import {errorAlert, timeSince} from '../../../util';
import {postRequest} from '../../../api';
import useAdminStore from '../../../store';
import {twMerge} from 'tailwind-merge';

interface JudgeRowProps {
    judge: Judge;
    idx: number;
    checked: boolean;
    handleCheckedChange: (e: React.ChangeEvent<HTMLInputElement>, idx: number) => void;
}

const JudgeRow = ({ judge, idx, checked, handleCheckedChange }: JudgeRowProps) => {
    const [popup, setPopup] = useState(false);
    const ref = useRef<HTMLDivElement>(null);
    const fetchJudges = useAdminStore((state) => state.fetchJudges);

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
            case 'hide':
                // Hide
                hideJudge();
                break;
        }

        setPopup(false);
    };

    const hideJudge = async () => {
        const res = await postRequest<YesNoResponse>(judge.active ? '/judge/hide' : '/judge/unhide', {id: judge.id});
        if (res.status === 200) {
            alert(`Judge account ${judge.active ? 'disabled' : 're-enabled'} successfully!`);
            await fetchJudges();
        } else {
            errorAlert(res);
        }
    };

    return (
        <>
            <tr
                key={idx}
                className={twMerge(
                    'border-t-2 border-backgroundDark duration-150',
                    checked ? 'bg-primary/20' : !judge.active ? 'bg-lightest' : 'bg-background'
                )}
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
                <td>{judge.name}</td>
                <td className="text-center">{judge.keycloak_user_id}</td>
                <td className="text-center">{judge.seen}</td>
                <td className="text-center">{judge.past_rankings ? judge.past_rankings.length : 0}</td>
                <td className="text-center">{timeSince(judge.last_activity)}</td>
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
                                {judge.active ? 'Restrict/Disable' : 'Unrestrict/Re-enable'}
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
        </>
    );
};

export default JudgeRow;
