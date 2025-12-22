import { createFileRoute } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import { useGETApiUsersPublicKey } from "@/lib/api/generated/api/api";
import { hexToNpub } from "@/lib/crypto";

export const Route = createFileRoute("/user/$pubkey")({
	component: UserProfilePage,
});

function UserProfilePage() {
	const { pubkey } = Route.useParams();

	const { data: response, isLoading, error } = useGETApiUsersPublicKey(pubkey);

	if (isLoading) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center">Loading user profile...</div>
			</div>
		);
	}

	if (error || !response || response.status !== 200) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center text-red-500">
					Error loading user profile
				</div>
			</div>
		);
	}

	const user = response.data;

	if (!user.created_at || !user.public_key) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center text-red-500">Invalid user data</div>
			</div>
		);
	}

	const createdDate = formatDistanceToNow(new Date(user.created_at), {
		addSuffix: true,
	});

	const npub = hexToNpub(user.public_key);

	return (
		<div className="container mx-auto p-8">
			<div className="max-w-2xl mx-auto">
				<h1 className="text-3xl font-bold mb-6">User Profile</h1>

				<div className="bg-muted rounded-lg p-6 space-y-4">
					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Public Key (npub)
						</div>
						<p className="font-mono text-sm break-all mt-1">{npub}</p>
					</div>

					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Verified by {window.location.hostname} admins
						</div>
						<p className="mt-1">
							{user.verified ? (
								<span className="text-green-500">✓ Verified</span>
							) : (
								<span className="text-muted-foreground">Not verified</span>
							)}
						</p>
					</div>

					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Member since
						</div>
						<p className="mt-1">{createdDate}</p>
					</div>
				</div>
			</div>
		</div>
	);
}
