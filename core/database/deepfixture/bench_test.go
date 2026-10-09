package deepfixture_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/deepfixture"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
)

func openDeep(b *testing.B) *deepfixture.Catalog {
	b.Helper()
	cat, err := deepfixture.Generate(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = cat.Cat.Close() })
	return cat
}

func BenchmarkRebuild(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := autoreconciler.Rebuild(db); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkObservationUpkeep(b *testing.B) {
	cat := openDeep(b)
	obs := cat.NameObservation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := observations.Update(cat.Cat, deepfixture.UserID, observations.Input{
			ID: obs.ID, SubjectID: obs.SubjectID, PropertyID: obs.PropertyID,
			Name: namevaluestest.Western("James Kenneth Robins"),
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProposeObituary(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := promotealign.Propose(db, cat.ObitSource, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProposeDeep(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := promotealign.Propose(db, cat.DeepSource, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDoneObituary(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		sourceID, subjects := cat.ObitSource, cat.ObitSubjects
		if i > 0 {
			sourceID, subjects, err = cat.Unpromoted("Obituary again", 6, 4)
			if err != nil {
				b.Fatal(err)
			}
		}
		_, rev, err := promotealign.Propose(db, sourceID, nil)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		if err := fileNew(cat, sourceID, rev, subjects); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDoneDeep(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		sourceID, subjects, err := cat.Unpromoted("Deep batch", 6, 4)
		if err != nil {
			b.Fatal(err)
		}
		_, rev, err := promotealign.Propose(db, sourceID, nil)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		if err := fileNew(cat, sourceID, rev, subjects); err != nil {
			b.Fatal(err)
		}
	}
}

func fileNew(cat *deepfixture.Catalog, sourceID []byte, rev int64, subjects [][]byte) error {
	rows := make([]promote.BatchRow, len(subjects))
	for i, id := range subjects {
		rows[i] = promote.BatchRow{SubjectID: id, Target: promote.TargetNew}
	}
	_, err := promote.SaveBatch(cat.Cat, deepfixture.UserID, promote.Batch{
		SourceID: sourceID, SeenRevision: rev, Rows: rows,
	})
	return err
}

func BenchmarkListComposition(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conclusionheaders.ListPersons(db); err != nil {
			b.Fatal(err)
		}
		if _, err := conclusionheaders.ListEvents(db); err != nil {
			b.Fatal(err)
		}
		if _, err := conclusionheaders.ListPlaces(db); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOneDetail(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conclusiondetails.ForEntity(db, cat.Person); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPlaceChain(b *testing.B) {
	cat := openDeep(b)
	db, err := cat.Cat.DB()
	if err != nil {
		b.Fatal(err)
	}
	y := 2026
	at := datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		names, err := conclusionheaders.ParentsAtDate(db, cat.Place, at)
		if err != nil {
			b.Fatal(err)
		}
		if len(names) < 4 {
			b.Fatalf("chain %v", names)
		}
	}
}
