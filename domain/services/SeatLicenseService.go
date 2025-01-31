package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
	vo "authz/domain/valueobjects"
)

// SeatLicenseService performs operations related to per-seat licensing
type SeatLicenseService struct {
	seats contracts.SeatLicenseRepository
	authz contracts.AccessRepository
}

// ModifySeats handles ModifySeatAssignmentEvents to assign and unassign seats
func (l *SeatLicenseService) ModifySeats(evt model.ModifySeatAssignmentEvent) error {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor); err != nil {
		return err
	}

	//TODO: consistency? Atm, if an error occurs part-way through, this will partially save.
	for _, principal := range evt.UnAssign {
		if err := l.seats.UnAssignSeat(principal, evt.Org.ID, evt.Service); err != nil {
			return err
		}
	}

	for _, principal := range evt.Assign {
		if err := l.seats.AssignSeat(principal, evt.Org.ID, evt.Service); err != nil {
			return err
		}
	}

	return nil
}

// GetLicense gets the License for the provided information
func (l *SeatLicenseService) GetLicense(evt model.GetLicenseEvent) (*model.License, error) {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor); err != nil {
		return nil, err
	}

	return l.seats.GetLicense(evt.OrgID, evt.ServiceID)
}

// GetAssignedSeats gets the subjects assigned to the given license
func (l *SeatLicenseService) GetAssignedSeats(evt model.GetLicenseEvent) ([]vo.SubjectID, error) {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor); err != nil {
		return nil, err
	}

	return l.seats.GetAssigned(evt.OrgID, evt.ServiceID)
}

// NewSeatLicenseService constructs a new SeatLicenseService
func NewSeatLicenseService(seats contracts.SeatLicenseRepository, authz contracts.AccessRepository) *SeatLicenseService {
	return &SeatLicenseService{seats: seats, authz: authz}
}

func (l *SeatLicenseService) ensureRequestorIsAuthorizedToManageLicenses(requestor vo.SubjectID) error {
	if !requestor.HasIdentity() {
		return model.ErrNotAuthenticated
	}

	authz, err := true, error(nil) //l.authz.CheckAccess(requestor, "manage_license", org.AsResource()) //Maybe on a per-service basis? //TODO: implement meta-authz
	if err != nil {
		return err
	}

	if !authz {
		return model.ErrNotAuthorized
	}

	return nil
}
