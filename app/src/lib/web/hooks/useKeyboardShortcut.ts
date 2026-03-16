import { useEffect, useRef } from "react";

export function useKeyboardShortcut(handler: (e: KeyboardEvent) => void) {
	const handlerRef = useRef(handler);

	// Update ref when handler changes
	useEffect(() => {
		handlerRef.current = handler;
	});

	// Add event listener only once
	useEffect(() => {
		const eventHandler = (e: KeyboardEvent) => handlerRef.current(e);
		document.addEventListener("keydown", eventHandler);
		return () => document.removeEventListener("keydown", eventHandler);
	}, []);
}
