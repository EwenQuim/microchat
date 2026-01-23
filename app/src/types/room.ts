import type { Room as APIRoom } from "@/lib/api/generated/openAPI.schemas";

/**
 * Client-side Room type that extends the API Room type with additional properties
 */
export interface Room extends APIRoom {
	/** Client-side property to track visited rooms */
	visited?: boolean;
}
